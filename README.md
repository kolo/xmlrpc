## Overview

xmlrpc is an implementation of client side part of XMLRPC protocol in Go language.

## Installation

To install xmlrpc package run `go get github.com/kolo/xmlrpc`. To use
it in application add `"github.com/kolo/xmlrpc"` string to `import`
statement.

## Usage

    client, _ := xmlrpc.NewClient("https://bugzilla.mozilla.org/xmlrpc.cgi")
    result := struct{
      Version string `xmlrpc:"version"`
    }{}
    client.Call("Bugzilla.version", nil, &result)
    fmt.Printf("Version: %s\n", result.Version) // Version: 4.2.7+


Optionally, you may provide the following options after the URL argument in NewClient:

* xmlrpc.BasicAuth(username, password string) - The credentials are passed in
the Authorization HTTP request header per
[RFC1945](https://tools.ietf.org/html/rfc1945#section-11.1).
* xmlrpc.Transport([http.RoundTripper](http://golang.org/pkg/net/http/#RoundTripper)) -
The RoundTripper interface can be used to get more control over connection
options. By default it initialized by http.DefaultTransport object.

### Arguments encoding

xmlrpc package supports encoding of native Go data types to method
arguments.

Data types encoding rules:

* int, int8, int16, int32, int64 encoded to int;
* float32, float64 encoded to double;
* bool encoded to boolean;
* string encoded to string;
* time.Time encoded to datetime.iso8601;
* xmlrpc.Base64 encoded to base64;
* slice decoded to array;

Structs decoded to struct by following rules:

* all public field become struct members;
* field name become member name;
* if field has xmlrpc tag, its value become member name.

Server method can accept few arguments, to handle this case there is
special approach to handle slice of empty interfaces (`[]interface{}`).
Each value of such slice encoded as separate argument.

### Result decoding

Result of remote function is decoded to native Go data type.

Data types decoding rules:

* int, i4 decoded to int, int8, int16, int32, int64;
* double decoded to float32, float64;
* boolean decoded to bool;
* string decoded to string;
* array decoded to slice;
* structs decoded following the rules described in previous section;
* datetime.iso8601 decoded as time.Time data type;
* base64 decoded to string.

## Implementation details

xmlrpc package contains clientCodec type, that implements
[rpc.ClientCodec](http://golang.org/pkg/net/rpc/#ClientCodec)
interface of [net/rpc](http://golang.org/pkg/net/rpc) package.

xmlrpc package works over HTTP protocol, but some internal functions
and data type were made public to make it easier to create another
implementation of xmlrpc that works over another protocol. To encode
request body there is EncodeMethodCall function. To decode server
response Response data type can be used.

## Contribution

Feel free to fork the project, submit pull requests, ask questions.

## Authors

Dmitry Maksimov (dmtmax@gmail.com)
