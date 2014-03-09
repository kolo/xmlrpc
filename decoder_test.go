package xmlrpc

import (
	"reflect"
	"testing"
	"time"
)

type book struct {
	Title  string
	Amount int
}

var unmarshalTests = []struct {
	value interface{}
	ptr   interface{}
	xml   string
}{
	{100, new(*int), "<value><int>100</int></value>"},
	{"Once upon a time", new(*string), "<value><string>Once upon a time</string></value>"},
	{"Mike & Mick <London, UK>", new(*string), "<value><string>Mike &amp; Mick &lt;London, UK&gt;</string></value>"},
	{"Once upon a time", new(*string), "<value>Once upon a time</value>"},
	{"T25jZSB1cG9uIGEgdGltZQ==", new(*string), "<value><base64>T25jZSB1cG9uIGEgdGltZQ==</base64></value>"},
	{true, new(*bool), "<value><boolean>1</boolean></value>"},
	{false, new(*bool), "<value><boolean>0</boolean></value>"},
	{12.134, new(*float32), "<value><double>12.134</double></value>"},
	{-12.134, new(*float32), "<value><double>-12.134</double></value>"},
	{time.Unix(1386622812, 0).UTC(), new(*time.Time), "<value><dateTime.iso8601>20131209T21:00:12</dateTime.iso8601></value>"},
	{[]int{1, 5, 7}, new(*[]int), "<value><array><data><value><int>1</int></value><value><int>5</int></value><value><int>7</int></value></data></array></value>"},
	{book{"War and Piece", 20}, new(*book), "<value><struct><member><name>Title</name><value><string>War and Piece</string></value></member><member><name>Amount</name><value><int>20</int></value></member></struct></value>"},
}

func Test_unmarshal(t *testing.T) {
	for _, tt := range unmarshalTests {
		v := reflect.New(reflect.TypeOf(tt.value))
		if err := unmarshal([]byte(tt.xml), v.Interface()); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		v = v.Elem()

		if v.Kind() == reflect.Slice {
			vv := reflect.ValueOf(tt.value)
			if vv.Len() != v.Len() {
				t.Fatalf("unmarshal error:\nexpected: %v\n     got: %v", tt.value, v.Interface())
			}
			for i := 0; i < v.Len(); i++ {
				if v.Index(i).Interface() != vv.Index(i).Interface() {
					t.Fatalf("unmarshal error:\nexpected: %v\n     got: %v", tt.value, v.Interface())
				}
			}
		} else {
			if v.Interface() != interface{}(tt.value) {
				t.Fatalf("unmarshal error:\nexpected: %v\n     got: %v", tt.value, v.Interface())
			}
		}
	}
}

func Test_unmarshalToNil(t *testing.T) {
	for _, tt := range unmarshalTests {
		if err := unmarshal([]byte(tt.xml), tt.ptr); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
	}
}

func Test_typeMismatchError(t *testing.T) {
	var r uint

	for _, tt := range unmarshalTests {
		var err error
		if err = unmarshal([]byte(tt.xml), &r); err == nil {
			t.Fatal("unmarshal error: expected error, but didn't get it")
		}

		if _, ok := err.(TypeMismatchError); !ok {
			t.Fatal("unmarshal error: expected type mistmatch error, but didn't get it")
		}
	}
}
