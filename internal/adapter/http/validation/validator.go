package http

import (
	"strings"
	"todoapp/internal/core/model/response"

	"github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ptbr_translations "github.com/go-playground/validator/v10/translations/pt_BR"
)

var (
	Validator  *validator.Validate
	Translator ut.Translator
)

func init() {
	Validator = validator.New(validator.WithRequiredStructEnabled())

	ptBR := pt_BR.New()
	uni := ut.New(ptBR, ptBR)

	var found bool
	Translator, found = uni.GetTranslator("pt_BR")

	if !found {
		panic("translator pt_BR not found")
	}

	if err := ptbr_translations.RegisterDefaultTranslations(Validator, Translator); err != nil {
		panic(err)
	}

	addCustomTranslations()
}

func addCustomTranslations() {
	Validator.RegisterTranslation("required", Translator, func(ut ut.Translator) error {
		return ut.Add("required", "{0} é obrigatório", true)

	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", getFieldName(fe.Field()))
		return t
	})

	Validator.RegisterTranslation("min", Translator, func(ut ut.Translator) error {
		return ut.Add("min", "{0} deve ter no mínimo {1} caracteres", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", getFieldName(fe.Field()), fe.Param())
		return t
	})

	Validator.RegisterTranslation("max", Translator, func(ut ut.Translator) error {
		return ut.Add("max", "{0} deve ter no máximo {1} caracteres", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", getFieldName(fe.Field()), fe.Param())
		return t
	})

	Validator.RegisterTranslation("email", Translator, func(ut ut.Translator) error {
		return ut.Add("email", "{0} deve ser um email válido", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", getFieldName(fe.Field()))
		return t
	})
}

func getFieldName(field string) string {
	fieldNames := map[string]string{
		"Title":       "Título",
		"Description": "Descrição",
		"Name":        "Nome",
		"Email":       "Email",
		"Password":    "Senha",
		"Status":      "Status",
		"Completed":   "Completado",
	}

	if name, exists := fieldNames[field]; exists {
		return name
	}

	return field
}

func FormatValidationErrors(err error) []response.ValidationError {
	var errors []response.ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, response.ValidationError{
				Field:   strings.ToLower(fieldError.Field()),
				Message: fieldError.Translate(Translator),
			})
		}
	}

	return errors
}
