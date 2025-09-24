package helper

import (
	"net/http"

	. "todoapp/internal/adapter/http/validation"
	"todoapp/internal/core/model/response"

	"github.com/gin-gonic/gin"
)

func SendSuccess(c *gin.Context, statusCode int, data any, message ...string) {
	response := response.SuccessResponse{
		Data: data,
	}

	if len(message) > 0 && message[0] != "" {
		response.Message = message[0]
	}

	c.JSON(statusCode, response)
}

func SendError(c *gin.Context, statusCode int, code string, errors []response.ValidationError, details ...any) {
	errorResponse := response.ErrorResponse{
		Error: response.ResponseError{
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
	errors := []response.ValidationError{
		{
			Field:   "server",
			Message: message,
		},
	}

	SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errors, details...)
}

func SendUnauthorizedError(c *gin.Context, message string) {
	errors := []response.ValidationError{
		{
			Field:   "auth",
			Message: message,
		},
	}

	SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", errors)
}

func SendBadRequestError(c *gin.Context, field string, message string) {
	errors := []response.ValidationError{
		{
			Field:   field,
			Message: message,
		},
	}

	SendError(c, http.StatusBadRequest, "BAD_REQUEST", errors)
}

func SendNotFoundError(c *gin.Context, message string) {
	errors := []response.ValidationError{
		{
			Field:   "resource",
			Message: message,
		},
	}

	SendError(c, http.StatusNotFound, "NOT_FOUND", errors)
}
