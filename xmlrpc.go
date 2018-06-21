package xmlrpc

import (
	"fmt"
)

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
	code interface{}
	err  string
}

// Error() method implements Error interface
func (e *xmlrpcError) Error() string {
	return fmt.Sprintf("error: \"%s\" code: %v", e.err, e.code)
}

// Base64 represents value in base64 encoding
type Base64 string
