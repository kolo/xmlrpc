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
)

type clientCodec struct {
    reader *bufio.Reader
    writer *bufio.Writer
    conn   io.Closer

    responseBody []byte
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

    return codec.writer.Flush()
}

func (codec *clientCodec) ReadResponseHeader(response *rpc.Response) error {
    httpResponse, err := http.ReadResponse(codec.reader, &http.Request{ Method: "POST" })

    if err != nil {
        return err
    }

    codec.responseBody, err = ioutil.ReadAll(httpResponse.Body)

    if err != nil {
        return err
    }

    httpResponse.Body.Close()

    return nil
}

func (codec *clientCodec) ReadResponseBody(body interface{}) (err error) {
    body, err = parseResponse(codec.responseBody)

    if err != nil {
        return err
    }

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

    codec := clientCodec{ reader: reader, writer: writer, conn: conn }
    return rpc.NewClientWithCodec(&codec), nil
}
