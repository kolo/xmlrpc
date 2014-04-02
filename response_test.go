package xmlrpc

import (
	"testing"
)

const faultRespXml = `
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

func Test_failedResponse(t *testing.T) {
	resp := NewResponse([]byte(faultRespXml))

	if !resp.Failed() {
		t.Fatal("Failed() error: expected true, got false")
	}

	if resp.Err() == nil {
		t.Fatal("Err() error: expected error, got nil")
	}

	err := resp.Err().(*xmlrpcError)
	if err.code != 410 && err.err != "You must log in before using this part of Bugzilla." {
		t.Fatal("Err() error: got wrong error")
	}
}
