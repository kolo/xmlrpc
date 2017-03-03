package xmlrpc

import (
	"testing"
	"time"
	"net/http"
)



func Test_CallWithoutArgs(t *testing.T) {
	clients, closeClients := getTestClients(t)
	defer closeClients()
	for _, client := range clients {
		var result time.Time
		if err := client.Call("service.time", nil, &result); err != nil {
			t.Fatalf("service.time call error: %v", err)
		}
	}
}

func Test_CallWithOneArg(t *testing.T) {
	clients, closeClients := getTestClients(t)
	defer closeClients()
	for _, client := range clients {
		var result string
		if err := client.Call("service.upcase", "xmlrpc", &result); err != nil {
			t.Fatalf("service.upcase call error: %v", err)
		}

		if result != "XMLRPC" {
			t.Fatalf("Unexpected result of service.upcase: %s != %s", "XMLRPC", result)
		}
	}
}

func Test_CallWithTwoArgs(t *testing.T) {
	clients, closeClients := getTestClients(t)
	defer closeClients()
	for _, client := range clients {
		var sum int
		if err := client.Call("service.sum", []interface{}{2, 3}, &sum); err != nil {
			t.Fatalf("service.sum call error: %v", err)
		}

		if sum != 5 {
			t.Fatalf("Unexpected result of service.sum: %d != %d", 5, sum)
		}
	}
}

func Test_TwoCalls(t *testing.T) {
	clients, closeClients := getTestClients(t)
	defer closeClients()
	for _, client := range clients {
		var upcase string
		if err := client.Call("service.upcase", "xmlrpc", &upcase); err != nil {
			t.Fatalf("service.upcase call error: %v", err)
		}

		var sum int
		if err := client.Call("service.sum", []interface{}{2, 3}, &sum); err != nil {
			t.Fatalf("service.sum call error: %v", err)
		}
	}
}

func Test_FailedCall(t *testing.T) {
	clients, closeClients := getTestClients(t)
	defer closeClients()
	for _, client := range clients {
		var result int
		if err := client.Call("service.error", nil, &result); err == nil {
			t.Fatal("expected service.error returns error, but it didn't")
		}
	}
}

func newTestClient(t *testing.T) *Client {
	client, err := NewClient("http://localhost:5001", nil)
	if err != nil {
		t.Fatalf("Can't create client: %v", err)
	}

	return client
}

func newTestClientHttpClient(t *testing.T) *Client {
	client, err := NewClientHttpClient("http://localhost:5001", http.DefaultClient)
	if err != nil {
		t.Fatalf("Can't create client from http.Client: %v", err)
	}

	return client
}

func getTestClients(t *testing.T) ([]*Client, func()) {
	clients := []*Client{}
	clients = append(clients, newTestClient(t))
	clients = append(clients, newTestClientHttpClient(t))
	return clients, func() {
		for _, client := range clients {
			client.Close()
		}
	}
}
