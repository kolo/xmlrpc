package xmlrpc

import (
	"testing"
	"time"
)

func Test_parseValue_dateTime(t *testing.T) {
	result, _ := parseValue([]byte("<value><dateTime.iso8601>2012091T16:12:04</dateTime.iso8601></value>"))
	time, _ := time.Parse(TIME_LAYOUT, "2012091716:12:04")
	assert_equal(t, time, result)
}

func Test_parseValue_Int(t *testing.T) {
	result, _ := parseValue([]byte("<value><int>410</int></value>"))

	var i int64
	i = 410

	assert_equal(t, i, result)
}

func Test_parseValue_String(t *testing.T) {
	result, _ := parseValue([]byte("<value><string>Hello, world!</string></value>"))
	assert_equal(t, "Hello, world!", result)
}

func Test_parseValue_Boolean(t *testing.T) {
	result, _ := parseValue([]byte("<value><boolean>1</boolean></value>"))
	assert_equal(t, true, result)
	result, _ = parseValue([]byte("<value><boolean>0</boolean></value>"))
	assert_equal(t, false, result)
}

func Test_parseValue_Double(t *testing.T) {
	result, _ := parseValue([]byte("<value><double>10.345</double></value>"))
	assert_equal(t, 10.345, result)
}
