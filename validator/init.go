package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/leijiru1994/go-sdk/util/phone"
	"github.com/leijiru1994/go-sdk/validator/option"
	"gopkg.in/go-playground/validator.v9"
	translations "gopkg.in/go-playground/validator.v9/translations/zh"
)

var (
	defaultTranslator ut.Translator
)

const (
	FieldTag4ErrorDisplay = "trans_display"
	FieldTag4Json         = "json"
	FieldTag4Form         = "form"
)

func Init(opts ...option.Option) (err error) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	zh := zh.New()
	uni := ut.New(zh)
	defaultTranslator, _ = uni.GetTranslator("zh")
	err = translations.RegisterDefaultTranslations(v, defaultTranslator)
	if err != nil {
		return
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		var name string
		for _, field := range []string{FieldTag4ErrorDisplay, FieldTag4Json, FieldTag4Form} {
			if tmp := strings.SplitN(fld.Tag.Get(field), ",", 2)[0]; tmp != "-" && tmp != "" {
				name = tmp
				break
			}
		}

		return name
	})

	for _, opt := range opts {
		err = v.RegisterValidation(opt.Tag, opt.ValidateFn)
		if err != nil {
			return
		}

		err = v.RegisterTranslation(
			opt.Tag,
			defaultTranslator,
			opt.RegisterFn,
			opt.TranslationFn,
		)
		if err != nil {
			return
		}
	}

	return
}

func ErrorTipAfterTranslate(err error) (tip string) {
	tip = err.Error()
	if d, ok := err.(validator.ValidationErrors); ok {
		trans := d.Translate(defaultTranslator)
		transList := make([]string, len(trans))

		var index int
		for _, v := range trans {
			transList[index] = v
			index++
		}

		tip = strings.Join(transList, ";")
	}

	return
}

func IsValidChinaPhone(fl validator.FieldLevel) bool {
	return phone.IsChinaMobile(fl.Field().String())
}
