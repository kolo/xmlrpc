package xmlrpc

import (
    "fmt"
)

// TIME_LAYOUT defines time template defined by iso8601, used to encode/decode time values.
const TIME_LAYOUT = "20060102T15:04:05"

// Struct presents hash type used in xmlprc requests and responses.
type Struct map[string]interface{}

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
    code int64
    message string
}

// Error() method implements Error interface
func (e *xmlrpcError) Error() string {
    return fmt.Sprintf("Error: \"%s\" Code: %d", e.message, e.code)
}
