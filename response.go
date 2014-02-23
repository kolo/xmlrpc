package xmlrpc

import (
	"fmt"
	"regexp"
)

var (
	faultRx = regexp.MustCompile(`<fault>(\s|\S)+</fault>`)
)

type Response struct {
	data []byte
}

func NewResponse(data []byte) *Response {
	return &Response{
		data: data,
	}
}

func (r *Response) Failed() bool {
	return faultRx.Match(r.data)
}

func (r *Response) Err() error {
	var valueXml []byte
	valueXml = getValueXml(r.data)

	value, err := parseValue(valueXml)
	faultDetails := value.(Struct)

	if err != nil {
		return err
	}

	return &(xmlrpcError{
		code:    fmt.Sprintf("%v", faultDetails["faultCode"]),
		message: faultDetails["faultString"].(string),
	})
}

func parseSuccessfulResponse(response []byte) (interface{}, error) {
	valueXml := getValueXml(response)
	return parseValue(valueXml)
}


func getValueXml(rawXml []byte) []byte {
	expr, _ := regexp.Compile(`<value>(\s|\S)+</value>`)
	return expr.Find(rawXml)

}
