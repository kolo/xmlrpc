package xmlrpc

import (
	"fmt"
	"reflect"
	"regexp"
)

var (
	faultRx = regexp.MustCompile(`<fault>(\s|\S)+</fault>`)
)

// FaultError is returned from the server when an invalid call is made
type FaultError struct {
	Code   int    `xmlrpc:"faultCode"`
	String string `xmlrpc:"faultString"`
}

// Error implements the error interface
func (e FaultError) Error() string {
	return fmt.Sprintf("Fault(%d): %s", e.Code, e.String)
}

type Response []byte

func (r Response) Err() error {
	if !faultRx.Match(r) {
		return nil
	}
	var fault FaultError
	if err := unmarshal(r, &fault); err != nil {
		return err
	}
	return fault
}

func (r Response) Unmarshal(v interface{}) error {
	if err := unmarshal(r, v); err != nil {
		return err
	}
	return nil
}

type responseMulticall []byte

func (r responseMulticall) Err() error {
	return Response(r).Err()
}

// tmp storage for multicall responses
type responsesError struct {
	err   *MulticallFault
	datas interface{} // slice/array of pointers
}

func (r responseMulticall) Unmarshal(v interface{}) error {
	out, ok := v.(*responsesError)
	if !ok {
		return fmt.Errorf("wrong type for destination")
	}

	switch ki := reflect.TypeOf(out.datas).Kind(); ki {
	case reflect.Array, reflect.Slice: // OK
	default:
		return fmt.Errorf("destination for multicall must be Array or Slice, got %s", ki)
	}
	outSlice := reflect.ValueOf(out.datas)

	parts, err := splitMulticall(r)
	if multicallErr, ok := err.(MulticallFault); ok {
		out.err = &multicallErr
		return nil
	} else if err != nil {
		return err
	}

	if outSlice.Len() != len(parts) {
		return fmt.Errorf("invalid number of return destinations : response needs %d, got %d", len(parts), outSlice.Len())
	}
	for i, xmlReturn := range parts {
		// pointer to one call's destination
		elem := outSlice.Index(i).Interface()

		// unmarshal expect a wrapping <value> tag
		xmlReturn = append(append([]byte("<value>"), xmlReturn...), "</value>"...)
		if err := unmarshal(xmlReturn, elem); err != nil {
			return fmt.Errorf("unmarshall number %d failed : %s", i, err.Error())
		}
	}
	return nil
}
