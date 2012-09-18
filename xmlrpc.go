package xmlrpc

import (
    "fmt"
    //"io/ioutil"
    "net/http"
    "regexp"
    "strings"
    "time"
)

//type Client struct {
//    url string
//    httpClient *http.Client
//}

//func NewClient(url string) (client *Client) {
//    client = &Client{ url: url, httpClient: &http.Client{} }
//    return
//}

// Call make request to XMLRPC server, parses its result and return as instance of Result interface.
//func (client *Client) Call(method string, params ...interface{}) (result interface{}, err error) {
//    var request *http.Request
//    var response *http.Response
//
//    request, err = buildRequest(client.url, method, params)
//
//    if err != nil {
//        return result, err
//    }
//
//    response, err = client.httpClient.Do(request)
//
//    if response != nil && err == nil {
//
//        var data []byte
//        data, err = ioutil.ReadAll(response.Body)
//        response.Body.Close()
//
//        result, err = parseResponse(data)
//    }
//
//    return
//}

func buildRequest(url, method string, params []interface{}) (request *http.Request, err error) {
    requestBody := buildRequestBody(method, params)
    request, err = http.NewRequest("POST", url, strings.NewReader(requestBody))

    request.Header.Set("Content-Type", "text/xml")
    request.Header.Set("Content-Length", fmt.Sprintf("%d", len(requestBody)))

    return
}

func buildRequestBody(method string, params []interface{}) (buffer string) {
    buffer += `<?xml version="1.0" encoding="UTF-8"?><methodCall>`
    buffer += fmt.Sprintf("<methodName>%s</methodName><params>", method)

    if params != nil {
        for _, value := range params {
            buffer += buildParamElement(value)
        }
    }

    buffer += "</params></methodCall>"

    return
}

func buildParamElement(value interface{}) (string) {
    return fmt.Sprintf("<param>%s</param>", buildValueElement(value))
}

func buildValueElement(value interface{}) (buffer string) {
    buffer = `<value>`

    switch v := value.(type) {
    case Struct:
        buffer += buildStructElement(v)
    case string:
        buffer += fmt.Sprintf("<string>%s</string>", v)
    case int, int8, int16, int32, int64:
        buffer += fmt.Sprintf("<int>%d</int>", v)
    case float32, float64:
        buffer += fmt.Sprintf("<double>%f</double>", v)
    case bool:
        buffer += buildBooleanElement(v)
    case time.Time:
        buffer += buildTimeElement(v)
    default:
        fmt.Errorf("Unsupported value type")
    }

    buffer += `</value>`

    return
}

func buildStructElement(param Struct) (buffer string) {
    buffer = `<struct>`

    for name, value := range param {
        buffer += fmt.Sprintf("<member><name>%s</name>", name)
        buffer += buildValueElement(value)
        buffer += `</member>`
    }

    buffer += `</struct>`

    return
}

func buildBooleanElement(value bool) (buffer string) {
    if value {
        buffer = `<boolean>1</boolean>`
    } else {
        buffer = `<boolean>0</boolean>`
    }

    return
}

func buildTimeElement(t time.Time) string {
    return fmt.Sprintf("<dateTime.iso8601>%d%d%dT%d:%d:%d</dateTime.iso8601>", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func parseResponse(response []byte) (result interface{}, err error) {
    if fault, _ := isFaultResponse(response); fault {
        return nil, parseFailedResponse(response)
    }

    return parseSuccessfulResponse(response)
}

// isFaultResponse checks whether response failed or not. Response defined as failed if it
// contains <fault>...</fault> section.
func isFaultResponse(response []byte) (bool, error) {
    fault := true
    faultRegexp, err := regexp.Compile(`<fault>(\s|\S)+</fault>`)

    if err == nil {
        fault = faultRegexp.Match(response)
    }

    return fault, err
}
