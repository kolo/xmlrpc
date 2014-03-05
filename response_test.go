package xmlrpc

import (
	"testing"
	"time"
)

const RESPONSE_BODY = `
<?xml version="1.0" ?>
<methodResponse>
  <params>
    <param>
      <value>
        <dateTime.iso8601>20120911T18:16:01</dateTime.iso8601>
      </value>
    </param>
  </params>
</methodResponse>`

const FAULT_RESPONSE_BODY = `
<?xml version="1.0" encoding="UTF-8"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultString</name>
          <value>
            <string>You must log in before using this part of Bugzilla.</string>
          </value>
        </member>
        <member>
          <name>faultCode</name>
          <value>
            <int>410</int>
          </value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`

func Test_parseResponseBody_SuccessfulResponse(t *testing.T) {
	resp := NewResponse([]byte(RESPONSE_BODY))
	var result time.Time
	if err := resp.Unmarshal(&result); err != nil {
		t.Fatalf("parseResponse raised error: %v", err)
	}

	time, _ := time.Parse("2006012T15:04:05", "20120911T18:16:01")
	if result != time {
		t.Fatalf("Response#Unmarshal error: expected %v = %v", result, time)
	}
}

func Test_parseResponseBody_FaultResponse(t *testing.T) {
	resp := NewResponse([]byte(FAULT_RESPONSE_BODY))
	assert_equal(t, true, resp.Failed())
	t.Log(resp.Err())
	assert_not_nil(t, resp.Err())
}

const XENAPI_RESPONSE = `
<methodResponse>
  <params>
    <param>
      <value>
        <struct>
          <member>
            <name>Status</name>
            <value>Success</value>
          </member>
          <member>
            <name>Value</name>
            <value>OpaqueRef:4b40767e-bc91-ca34-7e11-0ca46bb6b3e0</value>
          </member>
        </struct>
      </value>
    </param>
  </params>
</methodResponse>
`

type XenApiResult struct {
	Value string
	Status string
}

func Test_parse_XenAPI_ResponseBody(t *testing.T) {
	resp := NewResponse([]byte(XENAPI_RESPONSE))

	result := &XenApiResult{}
	err := resp.Unmarshal(result)

	expected := map[string]string{
		"Value":  "OpaqueRef:4b40767e-bc91-ca34-7e11-0ca46bb6b3e0",
		"Status": "Success",
	}

	t.Log(err)

	assert_nil(t, err)
	assert_equal(t, result.Value, expected["Value"])
	assert_equal(t, result.Status, expected["Status"])
}
