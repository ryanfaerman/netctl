package services

import (
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/leebenson/conform"
)

type validation struct {
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	trans    ut.Translator
}

// Validation provides validation services
var Validation validation

func init() {
	en := en.New()
	Validation.uni = ut.New(en, en)

	Validation.trans, _ = Validation.uni.GetTranslator("en")

	Validation.validate = validator.New(validator.WithRequiredStructEnabled())
	en_translations.RegisterDefaultTranslations(Validation.validate, Validation.trans)
}

type ValidationError map[string]string

func (e ValidationError) Error() string {
	var b strings.Builder

	for k, v := range e {
		b.WriteString(fmt.Sprintf("Key: '%s', Error: '%s'\n", k, v))
	}

	return b.String()
}

// Apply executes the validations defined on the given struct. The validations are
// defined as struct tags. For a full listing of available validations
// see: https://github.com/go-playground/validator
func (v validation) Apply(m any) error {
	output := make(map[string]string)

	conform.Strings(m)

	err := v.validate.Struct(m)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if ok {
			// We can't use the validator's translation because it doesn't
			// handle our custom valiidation tag in a way that we want. Specifically,
			// we want the Netcheckin.Callsign.AsHeard field name to be Callsign
			// rather than AsHeard. So we have to do this manually.
			for _, e := range errs {
				switch e.Namespace() {
				case "NetCheckin.Callsign.AsHeard":
					switch e.Tag() {
					case "eq=|alphanum":
						output[e.Namespace()] = "Callsign must be alphanumeric"
					}
				}
			}

			for field, e := range errs.Translate(v.trans) {
				switch field {
				case "NetCheckin.Callsign.AsHeard":
					// We already handled this above, so we want the noop
				default:
					output[field] = e
				}
			}
		}
		return ValidationError(output)
	}
	return nil
}
