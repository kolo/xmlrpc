package xmlrpc

import (
	"fmt"
)

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
	code int
	err  string
}

// Error() method implements Error interface
func (e *xmlrpcError) Error() string {
	return fmt.Sprintf("Error: \"%s\" Code: %d", e.err, e.code)
}
