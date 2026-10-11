package http

import (
	"clinic-vet-api/internal/shared/errors"
	"time"
)

// @Description Standardized API response structure.
type APIResponse struct {
	// Indicates whether the request was successful.
	Success bool `json:"success"`
	// The data payload of the response for successful requests.
	Data any `json:"data,omitempty"`
	// A message providing additional context about the response.
	Message string `json:"message,omitempty"`
	// Details of the error if the request was not successful.
	Error *ErrorInfo `json:"error,omitempty"`
	// Additional metadata for the response, such as pagination info.
	Meta any `json:"meta,omitempty"`
	// The timestamp when the response was generated.
	Timestamp time.Time `json:"timestamp"`
	// A unique identifier for the request.
	RequestID string `json:"request_id,omitempty"`
}

// @Description Detailed information about an error.
type ErrorInfo struct {
	// The error code.
	Code string `json:"code"`
	// The type of the error.
	Type string `json:"type"`
	// A descriptive error message.
	Message string `json:"message"`
	// A map of specific error details, often used for validation errors.
	Details map[string]string `json:"details,omitempty"`
	// The error stack trace (for development purposes).
	Stack string `json:"stack,omitempty"` // Develop Only
}

func (r *APIResponse) SuccessRequest(data any, message string) {
	r.Success = true
	r.Data = data
	r.Message = message
	r.Timestamp = time.Now()
}

func (r *APIResponse) SuccessWithMeta(data any, meta any, message string) *APIResponse {
	r.Success = true
	r.Message = message
	r.Data = data
	r.Meta = meta
	r.Timestamp = time.Now()
	return r
}

func (r *APIResponse) ErrorRequest(err error) *APIResponse {
	errorInfo := &ErrorInfo{
		Code:    "INTERNAL_SERVER_ERROR",
		Type:    "application",
		Message: "An unexpected error occurred",
	}

	switch e := err.(type) {
	case errors.BaseApplicationError: // Or CASE INFRA
		errorInfo.Code = e.ErrorCode()
		errorInfo.Type = e.ErrorType()
		errorInfo.Message = e.Error()
		errorInfo.Details = e.DetailMap()
	default:
		errorInfo.Message = err.Error()
	}

	return &APIResponse{
		Success:   false,
		Error:     errorInfo,
		Timestamp: time.Now(),
	}
}
