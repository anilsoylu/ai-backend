package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a custom error type
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e AppError) Error() string {
	return fmt.Sprintf("error code: %d, message: %s", e.Code, e.Message)
}

// ErrorResponse represents the structure of the error response
type ErrorResponse struct {
	Error AppError `json:"error"`
}

// ErrorHandler is a middleware for handling errors globally
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Check if it's our custom error type
			if appErr, ok := err.Err.(AppError); ok {
				c.JSON(appErr.Code, ErrorResponse{Error: appErr})
				return
			}

			// Handle unknown errors
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: AppError{
					Code:    http.StatusInternalServerError,
					Message: "Internal Server Error",
				},
			})
			return
		}
	}
}

// NewAppError creates a new AppError instance
func NewAppError(code int, message string) AppError {
	return AppError{
		Code:    code,
		Message: message,
	}
} 