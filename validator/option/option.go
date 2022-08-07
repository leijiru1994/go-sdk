package option

import (
	"gopkg.in/go-playground/validator.v9"
)

type Option struct {
	Tag           string
	ValidateFn    validator.Func
	RegisterFn    validator.RegisterTranslationsFunc
	TranslationFn validator.TranslationFunc
}
