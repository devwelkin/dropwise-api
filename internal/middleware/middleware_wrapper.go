package middleware

import (
	"net/http"
)

// Middleware defines a function that wraps an http.HandlerFunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Chain applies multiple middleware to a handler in the specified order
// The first middleware in the list will be the outermost wrapper
func Chain(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	// Start with the original handler
	result := handler

	// Apply each middleware in reverse order
	// This ensures the first middleware in the list is the outermost wrapper
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}

	return result
}

// ApplyMiddleware is a helper function to apply a single middleware to a handler
func ApplyMiddleware(handler http.HandlerFunc, middleware Middleware) http.HandlerFunc {
	return middleware(handler)
}
