package port

import "todoapp/internal/core/model/response"

type Validator interface {
	ValidateStruct(s interface{}) error
	FormatValidationErrors(err error) []response.ValidationError
}
