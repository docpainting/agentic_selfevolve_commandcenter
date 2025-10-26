package jsonrpc

import (
	"encoding/json"
	"fmt"
)

const Version = "2.0"

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// NewRequest creates a new JSON-RPC 2.0 request
func NewRequest(id interface{}, method string, params map[string]interface{}) *Request {
	return &Request{
		JSONRPC: Version,
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewNotification creates a new JSON-RPC 2.0 notification (no ID)
func NewNotification(method string, params map[string]interface{}) *Request {
	return &Request{
		JSONRPC: Version,
		Method:  method,
		Params:  params,
	}
}

// NewResponse creates a successful JSON-RPC 2.0 response
func NewResponse(id interface{}, result interface{}) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC 2.0 response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// IsNotification checks if the request is a notification
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Validate validates the JSON-RPC 2.0 request
func (r *Request) Validate() error {
	if r.JSONRPC != Version {
		return fmt.Errorf("invalid jsonrpc version: %s", r.JSONRPC)
	}
	if r.Method == "" {
		return fmt.Errorf("method is required")
	}
	return nil
}

// Marshal marshals the request to JSON
func (r *Request) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Marshal marshals the response to JSON
func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// UnmarshalRequest unmarshals JSON to a request
func UnmarshalRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// UnmarshalResponse unmarshals JSON to a response
func UnmarshalResponse(data []byte) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Handler is a function that handles a JSON-RPC method
type Handler func(params map[string]interface{}) (interface{}, error)

// Router routes JSON-RPC requests to handlers
type Router struct {
	handlers map[string]Handler
}

// NewRouter creates a new JSON-RPC router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

// Register registers a handler for a method
func (r *Router) Register(method string, handler Handler) {
	r.handlers[method] = handler
}

// Handle handles a JSON-RPC request
func (r *Router) Handle(req *Request) *Response {
	if err := req.Validate(); err != nil {
		return NewErrorResponse(req.ID, InvalidRequest, err.Error(), nil)
	}

	handler, exists := r.handlers[req.Method]
	if !exists {
		return NewErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("method not found: %s", req.Method), nil)
	}

	result, err := handler(req.Params)
	if err != nil {
		return NewErrorResponse(req.ID, InternalError, err.Error(), nil)
	}

	// Don't send response for notifications
	if req.IsNotification() {
		return nil
	}

	return NewResponse(req.ID, result)
}

// HandleBytes handles a JSON-RPC request from bytes
func (r *Router) HandleBytes(data []byte) ([]byte, error) {
	req, err := UnmarshalRequest(data)
	if err != nil {
		resp := NewErrorResponse(nil, ParseError, "parse error", nil)
		return resp.Marshal()
	}

	resp := r.Handle(req)
	if resp == nil {
		// Notification - no response
		return nil, nil
	}

	return resp.Marshal()
}

