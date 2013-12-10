## Overview

xmlrpc is an implementation of XMLRPC protocol in Go language. Currently
it implements only client side part of XMLRPC.

## Installation

To install xmlrpc package run `go get github.com/kolo/xmlrpc`. To use
it in application add `"github.com/kolo/xmlrpc"` string to `import`
statement.

## Usage

    client, _ := xmlrpc.NewClient("https://bugzilla.mozilla.org/xmlrpc.cgi", nil)
    result := xmlrpc.Struct{}
    client.Call("Bugzilla.version", nil, &result)
    fmt.Printf("Version: %s\n", result["version"]) // Version: 4.0.8+

Second argument in is [http.Transport](http://golang.org/pkg/net/http/#Transport)
object, it can be used to get more control over connection options.

## Implementation details

xmlrpc package contains clientCodec type, that implements [rpc.ClientCodec](http://golang.org/pkg/net/rpc/#ClientCodec)
interface of [net/rpc](http://golang.org/pkg/net/rpc) package.

## Contribution

Feel free to fork the project, submit pull requests, ask questions.

## Authors

Dmitry Maksimov (dmtmax@gmail.com)
