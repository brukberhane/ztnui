package api

import (
	"fmt"
	"net/http"
	"strings"
)

// APIError represents a non-success HTTP response from the ZeroTier API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("zerotier api %s %s: %d %s", e.Method, e.Path, e.StatusCode, e.Body)
	}
	return fmt.Sprintf("zerotier api %s %s: %d", e.Method, e.Path, e.StatusCode)
}

// IsUnauthorized reports whether the error is a 401 response.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if ok := asAPIError(err, &apiErr); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsNotFound reports whether the error is a 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if ok := asAPIError(err, &apiErr); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsControllerPath reports whether the error is for a controller API path.
func IsControllerPath(err error) bool {
	var apiErr *APIError
	if ok := asAPIError(err, &apiErr); ok {
		return strings.HasPrefix(apiErr.Path, "/controller")
	}
	return false
}

func asAPIError(err error, target **APIError) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		*target = apiErr
		return true
	}
	return false
}
