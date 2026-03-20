package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/dealance/shared/domain/errors"
)

// Response is the standard JSON envelope for all API responses.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorBody is the error payload in the response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta holds pagination metadata.
type Meta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Count      int    `json:"count"`
}

// OK sends a successful response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// OKPaginated sends a successful response with data and pagination meta.
func OKPaginated(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

// NoContent sends a 204 response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response derived from an AppError.
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, Response{
			Success: false,
			Error: &ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		c.Abort()
		return
	}

	// Fallback for unexpected errors — never expose internals
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
		},
	})
	c.Abort()
}
