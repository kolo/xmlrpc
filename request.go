package xmlrpc

import (
	"bytes"
	"fmt"
	"net/http"
)

func NewRequest(url string, method string, args interface{}) (*http.Request, error) {
	body, err := EncodeRequest(method, args)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "text/xml")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	return request, nil
}

func EncodeRequest(method string, args interface{}) ([]byte, error) {
	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)
	var err error

	if _, err = fmt.Fprintf(buf, "<methodCall><methodName>%s</methodName>", method); err != nil {
		return nil, err
	}

	if args != nil {
		if _, err = fmt.Fprintf(buf, "<params>"); err != nil {
			return nil, err
		}

		var t []interface{}
		var ok bool
		if t, ok = args.([]interface{}); !ok {
			t = []interface{}{args}
		}

		for _, arg := range t {
			b, err := marshal(arg)
			if err != nil {
				return nil, err
			}

			if _, err = fmt.Fprintf(buf, "<param>%s</param>", string(b)); err != nil {
				return nil, err
			}
		}

		if _, err = fmt.Fprintf(buf, "</params>"); err != nil {
			return nil, err
		}
	}

	if _, err = buf.WriteString("</methodCall>"); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
