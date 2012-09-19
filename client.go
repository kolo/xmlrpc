package xmlrpc

import (
    "io/ioutil"
    "net/http"
    "net/rpc"
    "reflect"
)

// clientCodec is rpc.ClientCodec interface implementation.
type clientCodec struct {
    // url presents url of xmlrpc service
    url string

    // httpClient works with HTTP protocol
    httpClient *http.Client

    // responses presents map of active requests. It is required to return request id, that
    // rpc.Client can mark them as done.
    responses map[uint64]*http.Response

    // responseBody holds response body of last request.
    responseBody []byte

    // ready presents channel, that is used to link request and it`s response.
    ready chan uint64
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, params interface{}) (err error) {
    httpRequest, err := newRequest(codec.url, request.ServiceMethod, params)

    if err != nil {
        return err
    }

    var httpResponse *http.Response
    httpResponse, err = codec.httpClient.Do(httpRequest)

    if err != nil {
        return err
    }

    codec.responses[request.Seq] = httpResponse
    codec.ready <- request.Seq

    return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) (err error) {
    seq := <-codec.ready
    httpResponse := codec.responses[seq]

    codec.responseBody, err = ioutil.ReadAll(httpResponse.Body)

    if err != nil {
        return err
    }

    httpResponse.Body.Close()

    response.Seq = seq
    delete(codec.responses, seq)

    return nil
}

func (codec *clientCodec) ReadResponseBody(body interface{}) (err error) {
    var result interface{}
    result, err = parseResponse(codec.responseBody)

    if err != nil {
        return err
    }

    v := reflect.ValueOf(body)

    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    v.Set(reflect.ValueOf(result))

    return nil
}

func (codec *clientCodec) Close() error {
    transport := codec.httpClient.Transport.(*http.Transport)
    transport.CloseIdleConnections()
    return nil
}

// NewClient returns instance of rpc.Client object, that is used to send request to xmlrpc service.
func NewClient(url string, transport *http.Transport) (*rpc.Client, error) {
    if transport == nil {
        transport = &http.Transport{}
    }

    httpClient := &http.Client{ Transport: transport }

    codec := clientCodec{
        url: url,
        httpClient: httpClient,
        ready: make(chan uint64),
        responses: make(map[uint64]*http.Response),
    }

    return rpc.NewClientWithCodec(&codec), nil
}
