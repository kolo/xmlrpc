package xmlrpc

import (
    "testing"
    "time"
)

// To make this test passed, xmlrpc server should be run.
//func Test_Call(t *testing.T) {
//    client := NewClient("http://localhost:5001")
//    result, err := client.Call("bugzilla.time")
//
//    if err != nil {
//        t.Error(err)
//    }
//
//    assert_not_nil(t, result)
//    assert_nil(t, err)
//}

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

func Test_parseResponseBody_SuccessfulResponse(t *testing.T) {
    result, err := parseResponse([]byte(RESPONSE_BODY))

    if err != nil {
        t.Fatalf("parseResponse raised error: %v", err)
    }

    time, _ := time.Parse("2006012T15:04:05", "20120911T18:16:01")
    assert_equal(t, time, result)
}

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

func Test_parseResponseBody_FaultResponse(t *testing.T) {
    _, err := parseResponse([]byte(FAULT_RESPONSE_BODY))
    assert_not_nil(t, err)
}

//func Test_Call_WithParams(t *testing.T) {
//    client := NewClient("http://localhost:5001")
//    result, err := client.Call("bugzilla.login", Struct{"username": "joesmith", "password": "secret"})
//
//    var id int64
//    id = 120
//
//    assert_nil(t, err)
//    assert_equal(t, id, result.(Struct)["id"])
//}

func assert_not_nil(t *testing.T, object interface{}) {
    if object == nil {
        t.Error("Expected object not to be a nil, but it is.")
    }
}

func assert_nil(t *testing.T, object interface{}) {
    if object != nil {
        t.Error("Expected object to be a nil, but it is not.")
    }
}

func assert_equal(t *testing.T, expected interface{}, actual interface{}) {
    if expected != actual {
        t.Error("Expected objects to be equal, but they are not.")
    }
}
