package validator

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	Errors []ValidationError
}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

func (v *Validator) RequiredString(value, field string) {
	v.Check(strings.TrimSpace(value) != "", field, "this field is required")
}

func (v *Validator) MinLength(value, field string, min int) {
	v.Check(utf8.RuneCountInString(value) >= min, field, "must be at least "+string(rune(min+'0'))+" characters")
}

func (v *Validator) MaxLength(value, field string, max int) {
	v.Check(utf8.RuneCountInString(value) <= max, field, "must be at most "+string(rune(max+'0'))+" characters")
}

func (v *Validator) ValidEmail(value, field string) {
	_, err := mail.ParseAddress(value)
	v.Check(err == nil, field, "must be a valid email address")
}

func (v *Validator) ValidUsername(value, field string) {
	v.Check(usernameRegex.MatchString(value), field, "must be 3-30 alphanumeric characters or underscores")
}

func (v *Validator) ValidPassword(value, field string) {
	v.Check(utf8.RuneCountInString(value) >= 8, field, "must be at least 8 characters")
	v.Check(utf8.RuneCountInString(value) <= 72, field, "must be at most 72 characters")
}

func (v *Validator) PositiveFloat(value float64, field string) {
	v.Check(value > 0, field, "must be a positive number")
}

func (v *Validator) InList(value, field string, list []string) {
	for _, item := range list {
		if value == item {
			return
		}
	}
	v.AddError(field, "must be one of: "+strings.Join(list, ", "))
}
