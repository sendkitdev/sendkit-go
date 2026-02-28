package sendkit

import "fmt"

// APIError represents an error response from the SendKit API.
type APIError struct {
	Name       string `json:"name"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("sendkit: %s (%d): %s", e.Name, e.StatusCode, e.Message)
}
