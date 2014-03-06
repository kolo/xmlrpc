package xmlrpc

import (
	"bytes"
	"fmt"
	"net/http"
)

func NewRequest(url string, method string, params ...interface{}) (*http.Request, error) {
	body, err := EncodeRequest(method, params)
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

func EncodeRequest(method string, params []interface{}) ([]byte, error) {
	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)
	var err error

	if _, err = fmt.Fprintf(buf, "<methodCall><methodName>%s</methodName><params>", method); err != nil {
		return nil, err
	}

	for _, p := range params {
		b, err := marshal(p)
		if err != nil {
			return nil, err
		}

		if _, err = fmt.Fprintf(buf, "<param>%v</param>", b); err != nil {
			return nil, err
		}
	}

	if _, err = buf.WriteString("</params></methodCall>"); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
