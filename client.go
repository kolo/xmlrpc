package xmlrpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
)

const (
	SystemMulticall = "system.multicall"
)

type Client struct {
	*rpc.Client
}

type Call struct {
	ServiceMethod string        `xmlrpc:"methodName"`
	Args          []interface{} `xmlrpc:"params"`
	Reply         interface{}   `xmlrpc-skip:"true"`
	Error         error         `xmlrpc-skip:"true"`
}

type pendingRequest struct {
	serviceMethod string
	httpResponse  *http.Response
}

func (client *Client) Multicall(calls []Call) error {
	return client.Call(SystemMulticall, calls, calls)
}

// clientCodec is rpc.ClientCodec interface implementation.
type clientCodec struct {
	// url presents url of xmlrpc service
	url string

	// httpClient works with HTTP protocol
	httpClient *http.Client

	// cookies stores cookies received on last request
	cookies []*http.Cookie

	// pending presents map of active requests. It is required to return request id, that
	// rpc.Client can mark them as done.
	pending map[uint64]*pendingRequest

	response *Response

	// ready presents channel, that is used to link request and it`s response.
	ready chan uint64
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, args interface{}) (err error) {
	httpRequest, err := NewRequest(codec.url, request.ServiceMethod, args)

	if codec.cookies != nil {
		for _, cookie := range codec.cookies {
			httpRequest.AddCookie(cookie)
		}
	}

	if err != nil {
		return err
	}

	var httpResponse *http.Response
	httpResponse, err = codec.httpClient.Do(httpRequest)

	if err != nil {
		return err
	}

	if codec.cookies == nil {
		codec.cookies = httpResponse.Cookies()
	}

	codec.pending[request.Seq] = &pendingRequest{request.ServiceMethod, httpResponse}
	codec.ready <- request.Seq

	return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) (err error) {
	seq := <-codec.ready
	req := codec.pending[seq]

	response.Seq = seq
	response.ServiceMethod = req.serviceMethod

	if req.httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("request error: bad status code - %d", req.httpResponse.StatusCode)
	}

	respData, err := ioutil.ReadAll(req.httpResponse.Body)

	if err != nil {
		return err
	}

	req.httpResponse.Body.Close()

	resp := NewResponse(respData, req.serviceMethod == SystemMulticall)

	if resp.Failed() {
		response.Error = fmt.Sprintf("%v", resp.Err())
	}

	codec.response = resp

	response.Seq = seq
	delete(codec.pending, seq)

	return nil
}

func (codec *clientCodec) ReadResponseBody(v interface{}) (err error) {
	if v == nil {
		return nil
	}

	fmt.Println(codec.response.multicall)
	if err = codec.response.Unmarshal(v); err != nil {
		return err
	}

	return nil
}

func (codec *clientCodec) Close() error {
	transport := codec.httpClient.Transport.(*http.Transport)
	transport.CloseIdleConnections()
	return nil
}

// NewClient returns instance of rpc.Client object, that is used to send request to xmlrpc service.
func NewClient(url string, transport http.RoundTripper) (*Client, error) {
	if transport == nil {
		transport = http.DefaultTransport
	}

	httpClient := &http.Client{Transport: transport}

	codec := clientCodec{
		url:        url,
		httpClient: httpClient,
		ready:      make(chan uint64),
		pending:    make(map[uint64]*pendingRequest),
	}

	return &Client{rpc.NewClientWithCodec(&codec)}, nil
}
