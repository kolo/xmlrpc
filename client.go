package xmlrpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/rpc"
	"net/url"
	"sync"
)

type Client struct {
	*rpc.Client
}

// clientCodec is rpc.ClientCodec interface implementation.
type clientCodec struct {
	// url presents url of xmlrpc service
	url *url.URL

	// httpClient works with HTTP protocol
	httpClient *http.Client

	// creds allows HTTP requests to include credentials
	creds Credentials

	// cookies stores cookies received on last request
	cookies http.CookieJar

	resMutex sync.Mutex // protects responses
	// responses presents map of active requests. It is required to return request id, that
	// rpc.Client can mark them as done.
	responses map[uint64]*http.Response

	response *Response

	// ready presents channel, that is used to link request and it`s response.
	ready chan uint64
}

// Credentials handle adding credentials to a request.
type Credentials interface {
	SetCredentials(request *http.Request)
}

// A ClientOption sets options.
type ClientOption func(*clientCodec)

// Transport specifics a specific RoundTripper instead of http.DefaultTransport.
func Transport(transport http.RoundTripper) ClientOption {
	return func(codec *clientCodec) {
		codec.httpClient = &http.Client{Transport: transport}
	}
}

type basicAuth struct {
	username string
	password string
}

// SetCredentials sets the basic auth username and password.
func (a *basicAuth) SetCredentials(request *http.Request) {
	request.SetBasicAuth(a.username, a.password)
}

// BasicAuth sets the credentials for basic authentication.
func BasicAuth(username, password string) ClientOption {
	return func(codec *clientCodec) {
		codec.creds = &basicAuth{
			username: username,
			password: password,
		}
	}
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, args interface{}) (err error) {
	httpRequest, err := NewRequest(codec.url.String(), request.ServiceMethod, args)

	if codec.creds != nil {
		codec.creds.SetCredentials(httpRequest)
	}

	if codec.cookies != nil {
		for _, cookie := range codec.cookies.Cookies(codec.url) {
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

	if codec.cookies != nil {
		codec.cookies.SetCookies(codec.url, httpResponse.Cookies())
	}

	codec.resMutex.Lock()
	codec.responses[request.Seq] = httpResponse
	codec.resMutex.Unlock()
	codec.ready <- request.Seq

	return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) (err error) {
	seq := <-codec.ready
	codec.resMutex.Lock()
	httpResponse := codec.responses[seq]
	delete(codec.responses, seq)
	codec.resMutex.Unlock()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return fmt.Errorf("request error: bad status code - %d", httpResponse.StatusCode)
	}

	respData, err := ioutil.ReadAll(httpResponse.Body)

	if err != nil {
		return err
	}

	httpResponse.Body.Close()

	resp := NewResponse(respData)

	if resp.Failed() {
		response.Error = fmt.Sprintf("%v", resp.Err())
	}

	codec.response = resp
	response.Seq = seq

	return nil
}

func (codec *clientCodec) ReadResponseBody(v interface{}) (err error) {
	if v == nil {
		return nil
	}

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
func NewClient(requrl string, opt ...ClientOption) (*Client, error) {
	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	u, err := url.Parse(requrl)

	if err != nil {
		return nil, err
	}

	codec := clientCodec{
		url:       u,
		ready:     make(chan uint64),
		responses: make(map[uint64]*http.Response),
		cookies:   jar,
	}

	for _, o := range opt {
		if o != nil {
			o(&codec)
		}
	}

	if codec.httpClient == nil {
		codec.httpClient = &http.Client{Transport: http.DefaultTransport}
	}

	return &Client{rpc.NewClientWithCodec(&codec)}, nil
}
