package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type encodeFunc func(reflect.Value) ([]byte, error)

func marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte{}, nil
	}

	val := reflect.ValueOf(v)
	return encodeValue(val)
}

func encodeValue(val reflect.Value) ([]byte, error) {
	var b []byte
	var err error

	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return b, nil
		}

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		switch val.Interface().(type) {
		case time.Time:
			t := val.Interface().(time.Time)
			b = []byte(fmt.Sprintf("<dateTime.iso8601>%s</dateTime.iso8601>", t.Format(iso8601)))
		default:
			b, err = encodeStruct(val)
		}
	case reflect.Slice:
		b, err = encodeSlice(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = []byte(fmt.Sprintf("<int>%s</int>", strconv.FormatInt(val.Int(), 10)))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b = []byte(fmt.Sprintf("<i4>%s</i4>", strconv.FormatUint(val.Uint(), 10)))
	case reflect.Float32, reflect.Float64:
		b = []byte(fmt.Sprintf("<double>%s</double>",
			strconv.FormatFloat(val.Float(), 'g', -1, val.Type().Bits())))
	case reflect.Bool:
		if val.Bool() {
			b = []byte("<boolean>1</boolean>")
		} else {
			b = []byte("<boolean>0</boolean>")
		}
	case reflect.String:
		var buf bytes.Buffer
		xml.Escape(&buf, []byte(val.String()))
		b = []byte(fmt.Sprintf("<string>%s</string>", buf.String()))
	default:
		return nil, fmt.Errorf("xmlrpc encode error: unsupported type")
	}

	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("<value>%s</value>", string(b))), nil
}

func encodeStruct(val reflect.Value) ([]byte, error) {
	var buf bytes.Buffer

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		name := f.Tag.Get("xmlrpc")
		if name == "" {
			name = f.Name
		}

		b, err := encodeValue(val.FieldByName(f.Name))
		if err != nil {
			return nil, err
		}

		if _, err = fmt.Fprintf(&buf, "<member><name>%s</name>%s</member>", name, string(b)); err != nil {
			return nil, err
		}
	}

	return []byte(fmt.Sprintf("<struct>%s</struct>", buf.String())), nil
}

func encodeSlice(val reflect.Value) ([]byte, error) {
	var buf bytes.Buffer

	for i := 0; i < val.Len(); i++ {
		b, err := encodeValue(val.Index(i))
		if err != nil {
			return nil, err
		}

		if _, err = buf.Write(b); err != nil {
			return nil, err
		}
	}

	return []byte(fmt.Sprintf("<array>%s</array>", buf.String())), nil
}
