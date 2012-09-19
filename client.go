package xmlrpc

import (
    "crypto/tls"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/rpc"
    "reflect"
    "strings"
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

func (codec *clientCodec) WriteRequest(request *rpc.Request, body interface{}) (err error) {
    if body == nil {
        body = []interface{}{}
    }
    requestBody := buildRequestBody(request.ServiceMethod, body.([]interface{}))
    httpRequest, err := http.NewRequest("POST", codec.url, strings.NewReader(requestBody))

    if err != nil {
        return err
    }

    httpRequest.Header.Set("Content-Type", "text/xml")
    httpRequest.Header.Set("Content-Lenght", fmt.Sprintf("%d", len(requestBody)))

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

func NewClient(url string) (*rpc.Client, error) {
    transport := &http.Transport{ TLSClientConfig: &tls.Config{ InsecureSkipVerify: true } }
    httpClient := &http.Client{ Transport: transport }

    codec := clientCodec{
        url: url,
        httpClient: httpClient,
        ready: make(chan uint64),
        responses: make(map[uint64]*http.Response),
    }
    return rpc.NewClientWithCodec(&codec), nil
}
