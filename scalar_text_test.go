package xmlrpc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestResponseScalarTextChunks(t *testing.T) {
	cases := []struct {
		name, body   string
		target, want interface{}
	}{
		{"string", "<string>first<![CDATA[second]]>third</string>", new(string), "firstsecondthird"},
		{"string interface", "<string><![CDATA[first]]>second<![CDATA[third]]></string>", new(interface{}), "firstsecondthird"},
		{"base64", "<base64>Zm9v<![CDATA[YmFy]]></base64>", new(string), "Zm9vYmFy"},
		{"int", "<int>1<![CDATA[23]]></int>", new(int), 123},
		{"i4", "<i4>1<!-- split -->23</i4>", new(int32), int32(123)},
		{"i8", "<i8>1<![CDATA[23]]></i8>", new(int64), int64(123)},
		{"boolean", "<boolean>t<![CDATA[rue]]></boolean>", new(bool), true},
		{"double", "<double>1<![CDATA[.25]]></double>", new(float64), 1.25},
		{"time", "<dateTime.iso8601>20260102T<![CDATA[03:04:05]]></dateTime.iso8601>", new(time.Time), time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		{"large", "<string>prefix<![CDATA[" + strings.Repeat("long text", 8192) + "]]>suffix</string>", new(string), "prefix" + strings.Repeat("long text", 8192) + "suffix"},
		{"ordinary string", "<string> unmodified &amp; text </string>", new(string), " unmodified & text "},
		{"ordinary int", "<int>123</int>", new(int), 123},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Response("<value>" + tc.body + "</value>").Unmarshal(tc.target); err != nil {
				t.Fatal(err)
			}
			got := reflect.ValueOf(tc.target).Elem().Interface()
			if !reflect.DeepEqual(got, tc.want) {
				if text, ok := got.(string); ok {
					t.Fatalf("decoded %d bytes, want %d", len(text), len(tc.want.(string)))
				}
				t.Fatalf("decoded %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestResponseScalarChunkBoundaries(t *testing.T) {
	var result struct {
		First  string `xmlrpc:"first"`
		Second int    `xmlrpc:"second"`
	}
	body := `<value><struct><member><name>fi<![CDATA[rst]]></name><value><string>A<![CDATA[B]]>C</string></value></member><member><name>second</name><value><int>42</int></value></member></struct></value>`
	if err := Response(body).Unmarshal(&result); err != nil {
		t.Fatal(err)
	}
	if result.First != "ABC" || result.Second != 42 {
		t.Errorf("decoded %+v", result)
	}
	var values []string
	if err := Response(`<value><array><data><value><string>A<![CDATA[B]]>C</string></value><value><string>D<![CDATA[E]]>F</string></value></data></array></value>`).Unmarshal(&values); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []string{"ABC", "DEF"}) {
		t.Errorf("decoded %q", values)
	}
	fault := Response(`<methodResponse><fault><value><struct><member><name>faultCode</name><value><int>42</int></value></member><member><name>faultString</name><value><string>first<![CDATA[second]]>third</string></value></member></struct></value></fault></methodResponse>`).Err()
	f, ok := fault.(FaultError)
	if !ok || f.Code != 42 || f.String != "firstsecondthird" {
		t.Fatalf("fault %#v", fault)
	}
}

func TestResponseScalarTextErrorsAndEmptyValues(t *testing.T) {
	for _, body := range []string{`<value><string>abc<![CDATA[def]]>`, `<value><string>abc<![CDATA[unterminated`} {
		var value string
		err := Response(body).Unmarshal(&value)
		var syntax *xml.SyntaxError
		if !errors.As(err, &syntax) {
			t.Fatalf("invalid XML error = %v", err)
		}
	}
	var wrong int
	if err := Response(`<value><string>A<![CDATA[B]]></string></value>`).Unmarshal(&wrong); err == nil {
		t.Fatal("missing type mismatch")
	}
	for _, body := range []string{`<value><int>12<![CDATA[x]]></int></value>`, `<value><int>12<!--split-->x</int></value>`} {
		var value int
		if err := Response(body).Unmarshal(&value); err == nil {
			t.Error("invalid numeric suffix was ignored")
		}
	}
	// Existing empty typed-value semantics leave the destination untouched.
	value := 17
	if err := Response(`<value><int/></value>`).Unmarshal(&value); err != nil || value != 17 {
		t.Fatalf("empty int = %d, %v", value, err)
	}
	text := "unchanged"
	if err := Response(`<value><string/></value>`).Unmarshal(&text); err != nil || text != "unchanged" {
		t.Fatalf("empty string = %q, %v", text, err)
	}
}

func TestClientScalarTextChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if _, err := ioutil.ReadAll(r.Body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<?xml version="1.0"?><methodResponse><params><param><value><string>first<![CDATA[second]]>third</string></value></param></params></methodResponse>`)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	var result string
	if err := client.Call("service.result", nil, &result); err != nil {
		t.Fatal(err)
	}
	if result != "firstsecondthird" {
		t.Fatalf("HTTP client decoded %q", result)
	}
}
