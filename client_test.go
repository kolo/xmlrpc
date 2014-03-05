package xmlrpc

import (
	"testing"
	"time"
)

func Test_Client_CallWithoutParams(t *testing.T) {
	client, err := NewClient("http://localhost:5001", nil)

	assert_nil(t, err)

	defer client.Close()

	var result = new(time.Time)
	if err = client.Call("bugzilla.time", nil, result); err != nil {
		t.Fatal(err)
	}

	assert_not_nil(t, result)
}

func Test_Client_CallWithParams(t *testing.T) {
	client, err := NewClient("http://localhost:5001", nil)

	assert_nil(t, err)

	defer client.Close()

	result := &struct{
		Id int `xmlrpc:"id"`
	}{}

	if err = client.Call("bugzilla.login", Struct{"username": "joe", "password": "secret"}, result); err != nil {
		t.Fatal(err)
	}

	assert_equal(t, 120, result.Id)
}

func Test_Client_TwoCalls(t *testing.T) {
	client, err := NewClient("http://localhost:5001", nil)
	assert_nil(t, err)

	defer client.Close()

	var result string
	err = client.Call("bugzilla.error", nil, &result)
	assert_not_nil(t, err)

	time_result := new(time.Time)
	if err = client.Call("bugzilla.time", nil, time_result); err != nil {
		t.Fatal(err)
	}

	assert_not_nil(t, result)
}
