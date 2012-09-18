package xmlrpc

import (
    "bufio"
    "fmt"
    "io"
    "io/ioutil"
    "net"
    "net/http"
    "net/rpc"
    "strings"
    "reflect"
)

type clientCodec struct {
    reader *bufio.Reader
    writer *bufio.Writer
    conn   io.Closer

    responseBody []byte

    ready chan uint64
    requests map[uint64]*http.Request
}

func (codec *clientCodec) WriteRequest(request *rpc.Request, body interface{}) (err error) {
    if body == nil {
        body = []interface{}{}
    }
    requestBody := buildRequestBody(request.ServiceMethod, body.([]interface{}))
    httpRequest, err := http.NewRequest("POST", "/", strings.NewReader(requestBody))

    if err != nil {
        return err
    }

    httpRequest.Header.Set("Content-Type", "text/xml")
    httpRequest.Header.Set("Content-Lenght", fmt.Sprintf("%d", len(requestBody)))

    err = httpRequest.Write(codec.writer)

    if err != nil {
        return err
    }

    if err = codec.writer.Flush(); err != nil {
        return err
    }

    codec.requests[request.Seq] = httpRequest
    codec.ready <- request.Seq

    return nil
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) error {
    seq := <-codec.ready

    httpResponse, err := http.ReadResponse(codec.reader, codec.requests[seq])

    if err != nil {
        return err
    }

    codec.responseBody, err = ioutil.ReadAll(httpResponse.Body)

    if err != nil {
        return err
    }

    httpResponse.Body.Close()

    response.Seq = seq
    delete(codec.requests, seq)

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
    return codec.conn.Close()
}

func NewClient(url string) (*rpc.Client, error) {
    conn, err := net.Dial("tcp", url)

    if err != nil {
        return nil, err
    }

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    codec := clientCodec{
        reader: reader,
        writer: writer,
        conn: conn,
        ready: make(chan uint64),
        requests: make(map[uint64]*http.Request),
    }
    return rpc.NewClientWithCodec(&codec), nil
}
