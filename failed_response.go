package xmlrpc

import (
    "fmt"
)

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
    code int64
    message string
}

func (e *xmlrpcError) Error() string {
    return fmt.Sprintf("Error: \"%s\" Code: %d", e.message, e.code)
}

func parseFailedResponse(response []byte) (err error) {
    var valueXml []byte
    valueXml = getValueXml(response)

    value, err := parseValue(valueXml)
    faultDetails := value.(Struct)

    if err != nil {
        return err
    }

    return &(xmlrpcError{code: faultDetails["faultCode"].(int64), message: faultDetails["faultString"].(string)})
}
