package xmlrpc

import (
    "testing"
    "time"
)

func Test_Client_Call(t *testing.T) {
    client, _ := NewClient("localhost:5001")
    defer client.Close()

    var result = new(time.Time)
    err := client.Call("bugzilla.time", nil, result)

    assert_nil(t, err)
    assert_not_nil(t, result)
}
