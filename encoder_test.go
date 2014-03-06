package xmlrpc

import (
	"time"
	"testing"
)

var marshalTests = []struct {
	value  interface{}
	result string
}{
	{100, "<value><int>100</int></value>"},
	{"Once upon a time", "<value><string>Once upon a time</string></value>"},
	{"Mike & Mick <London, UK>", "<value><string>Mike &amp; Mick &lt;London, UK&gt;</string></value>"},
	{true, "<value><boolean>1</boolean></value>"},
	{false, "<value><boolean>0</boolean></value>"},
	{12.134, "<value><double>12.134</double></value>"},
	{-12.134, "<value><double>-12.134</double></value>"},
	{time.Unix(1386622812, 0).UTC(), "<value><dateTime.iso8601>20131209T21:00:12</dateTime.iso8601></value>"},
	{[]interface{}{1,"one"}, "<value><array><value><int>1</int></value><value><string>one</string></value></array></value>"},
	{&struct{
		Title string
		Amount int
	}{"War and Piece", 20}, "<value><struct><member><name>Title</name><value><string>War and Piece</string></value></member><member><name>Amount</name><value><int>20</int></value></member></struct></value>"},
}

func Test_marshal(t *testing.T) {
	for _, tt := range marshalTests {
		b, err := marshal(tt.value)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		if string(b) != tt.result {
			t.Fatalf("marshal error:\nexpected: %s\n     got: %s", tt.result, string(b))
		}

	}
}
