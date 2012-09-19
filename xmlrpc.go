package xmlrpc

import (
    "fmt"
)

// Struct presents hash type used in xmlprc requests and responses.
type Struct map[string]interface{}

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
    code string
    message string
}

// Error() method implements Error interface
func (e *xmlrpcError) Error() string {
    return fmt.Sprintf("Error: \"%s\" Code: %s", e.message, e.code)
}
