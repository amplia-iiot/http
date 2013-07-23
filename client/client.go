// Package gorilla/http/client contains the lower level HTTP client implementation.
//
// Package client is divided into two layers. The upper layer is the client.Client layer
// which handles the vaguaries of http implementations. The lower layer, client.Conn is
// concerned with encoding and decoding the HTTP data stream on the wire.
//
// The interface presented by client.Client is very powerful. It is not expected that normal
// consumers of HTTP services will need to operate at this level and should instead user the
// higher level interfaces in the gorilla/http package.
package client

import (
	"errors"
	"fmt"
	"io"
)

// Header represents a HTTP header.
type Header struct {
	Key   string
	Value string
}

// Request represents a complete HTTP request.
type Request struct {
	Method  string
	URI     string
	Version string

	Headers []Header

	Body io.Reader
}

// Client represents a single connection to a http server. Client obeys KeepAlive conditions for
// HTTP but connection pooling is expected to be handled at a higher layer.
type Client struct {
	*Conn
}

// SendRequest marshalls a HTTP request to the wire.
func (c *Client) SendRequest(req *Request) error {
	if err := c.WriteRequestLine(req.Method, req.URI, req.Version); err != nil {
		return err
	}
	for _, h := range req.Headers {
		if err := c.WriteHeader(h.Key, h.Value); err != nil {
			return err
		}
	}
	if err := c.StartBody(); err != nil {
		return err
	}
	if req.Body != nil {
		if err := c.WriteBody(req.Body); err != nil {
			return err
		}
	}
	return nil
}

// Status represents an HTTP status code.
type Status struct {
	Code    int
	Message string
}

func (s Status) String() string { return fmt.Sprintf("%d %s", s.Code, s.Message) }

var invalidStatus Status

// ReadResponse unmarshalls a HTTP response.
func (c *Client) ReadResponse() (Status, []Header, io.Reader, error) {
	_, code, msg, err := c.ReadStatusLine()
	var headers []Header
	if err != nil {
		return invalidStatus, headers, nil, fmt.Errorf("ReadStatusLine: %v", err)
	}
	status := Status{code, msg}
	for {
		var key, value string
		var done bool
		key, value, done, err = c.ReadHeader()
		if err != nil || done {
			break
		}
		if key == "" || value == "" {
			err = errors.New("invalid header")
			break
		}
		headers = append(headers, Header{key, value})
	}
	return status, headers, c.ReadBody(), err
}
