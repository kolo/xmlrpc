package xmlrpc

import (
	"encoding/xml"
	"testing"
)

func Test_buildValueElement_EscapesString(t *testing.T) {
	escaped := buildValueElement("Johnson & Johnson > 1")
	assert_equal(t, "<value><string>Johnson &amp; Johnson &gt; 1</string></value>", escaped)
}

const PARAMS_REQUEST = `
<?xml version="1.0" encoding="UTF-8"?>
<methodCall>
  <methodName>method</methodName>
  <params>
    <param>
      <value>
        <string>user</string>
      </value>
    </param>
    <param>
      <value>
        <string>pass</string>
      </value>
    </param>
  </params>
</methodCall>
`

type MethodCall struct {
	XMLName    xml.Name   `xml:"methodCall"`
	MethodName string     `xml:"methodName"`
	Params     []ParamsTy `xml:"params>param"`
}

type ParamsTy struct {
	Value string `xml:"value>string"`
}

func Test_buildRequestBody(t *testing.T) {
	var xml_expected MethodCall
	xml.Unmarshal([]byte(PARAMS_REQUEST), &xml_expected)

	params := Params{Params: []interface{}{"user", "pass"}}
	request := buildRequestBody("method", []interface{}{params})

	var xml_request MethodCall
	xml.Unmarshal([]byte(request), &xml_request)

	assert_equal(t, xml_request.MethodName, xml_expected.MethodName)
	assert_equal(t, xml_request.Params[0], xml_expected.Params[0])
	assert_equal(t, xml_request.Params[1], xml_expected.Params[1])
}

func Test_buildArrayElement(t *testing.T) {
	a := []interface{}{1}
	var arrayElement = buildArrayElement(a)

	assert_equal(t, arrayElement, "<array><data><value><int>1</int></value></data></array>")
}
