package xmlrpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/rpc"
	"net/url"
	"sync"
)

const multicallMethod = "system.multicall"

type Client struct {
	*rpc.Client
}

// Multicall performs a multicall request.
// `args` should be constructed with `NewMulticallArg`
// and `outs` must be an array/slice of pointers.
// If the response contains at least one fault,
// the first is returned.
func (c Client) Multicall(calls []MulticallArg, outs ...interface{}) error {
	if len(calls) != len(outs) {
		return errors.New("lengths of calls and responses are not matching")
	}
	return c.Call(multicallMethod, calls, outs)
}

// store the method name as well
// used to special-case multicall
type responseMethod struct {
	response *http.Response
	method   string
}

// clientCodec is rpc.ClientCodec interface implementation.
type clientCodec struct {
	// url presents url of xmlrpc service
	url *url.URL

	// httpClient works with HTTP protocol
	httpClient *http.Client

	// cookies stores cookies received on last request
	cookies http.CookieJar

	// responses presents map of active requests. It is required to return request id, that
	// rpc.Client can mark them as done.
	responses map[uint64]responseMethod
	mutex     sync.Mutex

	response interface {
		Err() error
		Unmarshal(v interface{}) error
	}

	// ready presents channel, that is used to link request and it`s response.
	ready chan uint64

	// close notifies codec is closed.
	close chan uint64
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, args interface{}) (err error) {
	httpRequest, err := NewRequest(codec.url.String(), request.ServiceMethod, args)

	if err != nil {
		return err
	}

	if codec.cookies != nil {
		for _, cookie := range codec.cookies.Cookies(codec.url) {
			httpRequest.AddCookie(cookie)
		}
	}

	var httpResponse *http.Response
	httpResponse, err = codec.httpClient.Do(httpRequest)

	if err != nil {
		return err
	}

	if codec.cookies != nil {
		codec.cookies.SetCookies(codec.url, httpResponse.Cookies())
	}

	codec.mutex.Lock()
	codec.responses[request.Seq] = responseMethod{response: httpResponse, method: request.ServiceMethod}
	codec.mutex.Unlock()

	codec.ready <- request.Seq

	return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) (err error) {
	var seq uint64
	select {
	case seq = <-codec.ready:
	case <-codec.close:
		return errors.New("codec is closed")
	}
	response.Seq = seq

	codec.mutex.Lock()
	resp := codec.responses[seq]
	httpResponse, method := resp.response, resp.method
	delete(codec.responses, seq)
	codec.mutex.Unlock()

	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		response.Error = fmt.Sprintf("request error: bad status code - %d", httpResponse.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		response.Error = err.Error()
		return nil
	}

	if method == multicallMethod {
		codec.response = ResponseMulticall(body)
	} else {
		codec.response = Response(body)
	}

	if err := codec.response.Err(); err != nil {
		response.Error = err.Error()
		return nil
	}

	return nil
}

func (codec *clientCodec) ReadResponseBody(v interface{}) (err error) {
	if v == nil {
		return nil
	}
	return codec.response.Unmarshal(v)
}

func (codec *clientCodec) Close() error {
	if transport, ok := codec.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	close(codec.close)

	return nil
}

// NewClient returns instance of rpc.Client object, that is used to send request to xmlrpc service.
func NewClient(requrl string, transport http.RoundTripper) (*Client, error) {
	if transport == nil {
		transport = http.DefaultTransport
	}

	httpClient := &http.Client{Transport: transport}

	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	u, err := url.Parse(requrl)

	if err != nil {
		return nil, err
	}

	codec := clientCodec{
		url:        u,
		httpClient: httpClient,
		close:      make(chan uint64),
		ready:      make(chan uint64),
		responses:  make(map[uint64]responseMethod),
		cookies:    jar,
	}

	return &Client{rpc.NewClientWithCodec(&codec)}, nil
}
