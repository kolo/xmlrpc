package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
)

// Types used for unmarshalling

// main fault should be already checked
type response struct {
	Name   xml.Name `xml:"methodResponse"`
	Params []param  `xml:"params>param"`
}

type param struct {
	Value value `xml:"value"`
}

type member struct {
	Name  string `xml:"name"`
	Value value  `xml:"value"`
}

type value struct {
	Array  []value  `xml:"array>data>value"` // used for returns
	Struct []member `xml:"struct>member"`    // used for fault
	String string   `xml:"string"`           // used for faults
	Int    string   `xml:"int"`              // used for faults
	Raw    []byte   `xml:",innerxml"`
}

// getFaultResponse converts faultValue to Fault.
func getFaultResponse(fault []member) FaultError {
	var (
		code int
		str  string
	)

	for _, field := range fault {
		if field.Name == "faultCode" {
			code, _ = strconv.Atoi(field.Value.Int)
		} else if field.Name == "faultString" {
			str = field.Value.String
			if str == "" {
				str = string(field.Value.Raw)
			}
		}
	}

	return FaultError{Code: code, String: str}
}

// MulticallFault tracks the position of the fault.
type MulticallFault struct {
	FaultError
	Index      int    // 0 based
	methodName string // for better message
}

func (m MulticallFault) Error() string {
	return fmt.Sprintf("fault in call %d (%s) : %s", m.Index, m.methodName, m.FaultError.Error())
}

// returns xml encoded chunks, one for each multicall response
// if there is (at least) one fault, returns the first one
// as error
func splitMulticall(xmlraw []byte) ([][]byte, error) {
	// Unmarshal raw XML into the temporal structure
	var ret response

	dec := xml.NewDecoder(bytes.NewReader(xmlraw))
	if CharsetReader != nil {
		dec.CharsetReader = CharsetReader
	}

	if err := dec.Decode(&ret); err != nil {
		return nil, err
	}
	if L := len(ret.Params); L != 1 {
		return nil, fmt.Errorf("unexpected number of arguments : got %d", L)
	}
	// multicall returns one array of values
	returns := ret.Params[0].Value.Array

	out := make([][]byte, len(returns))
	for i, oneReturn := range returns {
		// multicall return are always wrapped in one-sized array
		// otherwise, it's a fault
		if len(oneReturn.Array) != 1 {
			fault := getFaultResponse(oneReturn.Struct)
			return nil, MulticallFault{Index: i, FaultError: fault}
		}
		// unwrap the value and store raw xml
		// to further process
		out[i] = oneReturn.Array[0].Raw
	}
	return out, nil
}

// MulticallArg stores one call
type MulticallArg struct {
	MethodName string        `xmlrpc:"methodName"`
	Params     []interface{} `xmlrpc:"params"` // 1-sized list containing the real arguments
}

func NewMulticallArg(method string, args interface{}) MulticallArg {
	return MulticallArg{MethodName: method, Params: []interface{}{args}}
}
