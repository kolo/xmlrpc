package xmlrpc

import (
	"regexp"
)

var (
	faultRx = regexp.MustCompile(`<fault>(\s|\S)+</fault>`)
)

type failedResponse struct {
	Code  int    `xmlrpc:"faultCode"`
	Error string `xmlrpc:"faultString"`
}

func (r *failedResponse) err() error {
	return &xmlrpcError{
		code: r.Code,
		err:  r.Error,
	}
}

type Response struct {
	data      []byte
	multicall bool
}

func NewResponse(data []byte, multicall bool) *Response {
	return &Response{
		data:      data,
		multicall: multicall,
	}
}

func (r *Response) Failed() bool {
	return faultRx.Match(r.data)
}

func (r *Response) Err() error {
	failedResp := new(failedResponse)
	if err := unmarshal(r.data, failedResp); err != nil {
		return err
	}

	return failedResp.err()
}

func (r *Response) Unmarshal(v interface{}) error {
	if r.multicall {
		if err := unmarshalMulticall(r.data, v); err != nil {
			return err
		}
	} else {
		if err := unmarshal(r.data, v); err != nil {
			return err
		}
	}

	return nil
}
