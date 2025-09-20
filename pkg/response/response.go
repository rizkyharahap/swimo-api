package response

type BaseResponse struct {
	Code    int         `json:"code,omitempty"`    // HTTP status code.
	Message string      `json:"message,omitempty"` // Message corresponding to the status code.
	Error   string      `json:"error,omitempty"`   // error message.
	Data    interface{} `json:"data,omitempty"`
}
