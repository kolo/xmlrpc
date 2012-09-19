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
    result, err := parseResponse([]byte(RESPONSE_BODY))

    if err != nil {
        t.Fatalf("parseResponse raised error: %v", err)
    }

    time, _ := time.Parse("2006012T15:04:05", "20120911T18:16:01")
    assert_equal(t, time, result)
}

func Test_parseResponseBody_FaultResponse(t *testing.T) {
    _, err := parseResponse([]byte(FAULT_RESPONSE_BODY))
    assert_not_nil(t, err)
}


