package shared

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseError struct {
	Code    string            `json:"code"`
	Errors  []ValidationError `json:"errors"`
	Details any               `json:"details,omitempty"`
}

type SuccessResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error ResponseError `json:"error"`
}

func SendSuccess(c *gin.Context, statusCode int, data any, message ...string) {
	response := SuccessResponse{
		Data: data,
	}

	if len(message) > 0 && message[0] != "" {
		response.Message = message[0]
	}

	c.JSON(statusCode, response)
}

func SendError(c *gin.Context, statusCode int, code string, errors []ValidationError, details ...any) {
	errorResponse := ErrorResponse{
		Error: ResponseError{
			Code:   code,
			Errors: errors,
		},
	}

	if len(details) > 0 {
		errorResponse.Error.Details = details[0]
	}

	c.JSON(statusCode, errorResponse)
}

func SendValidationError(c *gin.Context, err error) {
	validationErrors := FormatValidationErrors(err)
	SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", validationErrors)
}

func SendInternalError(c *gin.Context, message string, details ...any) {
	errors := []ValidationError{
		{
			Field:   "server",
			Message: message,
		},
	}

	SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errors, details...)
}

func SendUnauthorizedError(c *gin.Context, message string) {
	errors := []ValidationError{
		{
			Field:   "auth",
			Message: message,
		},
	}

	SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", errors)
}

func SendBadRequestError(c *gin.Context, field string, message string) {
	errors := []ValidationError{
		{
			Field:   field,
			Message: message,
		},
	}

	SendError(c, http.StatusBadRequest, "BAD_REQUEST", errors)
}

func SendNotFoundError(c *gin.Context, message string) {
	errors := []ValidationError{
		{
			Field:   "resource",
			Message: message,
		},
	}

	SendError(c, http.StatusNotFound, "NOT_FOUND", errors)
}
