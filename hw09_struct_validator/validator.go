package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var (
	ErrMaxExceeded   = errors.New("значение поля превышаем максимально допустимое значение")
	ErrMinNotReached = errors.New("значение поля меньше минимально допустимого значения")
	ErrNotInSet      = errors.New("значение поля не входит в допустимое множество значений")
	ErrInvalidLength = errors.New("неверная длина строки")
	ErrNoMatchRegexp = errors.New("значение поля не соответствует регулярному выражению")
	ErrInvalidStruct = errors.New("значение не является стуктурой")

	ErrInvalidRuleFormat = errors.New("некорректный формат правила")
	ErrUnknownRule       = errors.New("неизвестное правило")
	ErrInvalidParam      = errors.New("некорректный параметр правила")
)

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("ошибки валидации:\n")
	for _, err := range v {
		sb.WriteString(fmt.Sprintf("- %s: %v\n", err.Field, err.Err))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	errors := make(ValidationErrors, 0, 10)

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%w: %s", ErrInvalidStruct, t.Kind())
	}
	val := reflect.ValueOf(v)

	for i := 0; i < t.NumField(); i++ {
		tagValue := t.Field(i).Tag.Get("validate")
		if tagValue == "" || tagValue == "-" {
			continue // пропускаем поле
		}

		field := val.Field(i)
		fieldName := t.Field(i).Name

		rules := strings.Split(tagValue, "|")
		for _, rule := range rules {
			parts := strings.SplitN(rule, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("%w: %s", ErrInvalidRuleFormat, rule)
			}

			ruleName := parts[0]
			ruleParam := parts[1]

			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

				err := validateInt(fieldName, field, ruleName, ruleParam, &errors)
				if err != nil {
					return err
				}

			case reflect.String:

				err := validateString(fieldName, field, ruleName, ruleParam, &errors)
				if err != nil {
					return err
				}

			case reflect.Slice:

				elementType := field.Type().Elem()
				switch elementType.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

					for j := 0; j < field.Len(); j++ {
						elem := field.Index(j)
						elemName := fmt.Sprintf("%s[%d]", fieldName, j)

						err := validateInt(elemName, elem, ruleName, ruleParam, &errors)
						if err != nil {
							return err
						}
					}

				case reflect.String:

					for j := 0; j < field.Len(); j++ {
						elem := field.Index(j)
						elemName := fmt.Sprintf("%s[%d]", fieldName, j)

						err := validateString(elemName, elem, ruleName, ruleParam, &errors)
						if err != nil {
							return err
						}
					}
				default:
					continue
				}

			default:
				continue
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func validateInt(fieldName string, field reflect.Value, ruleName string, ruleParam string, errors *ValidationErrors) error {
	switch ruleName {
	case "max":
		return validateMax(fieldName, field.Int(), ruleParam, errors)
	case "min":
		return validateMin(fieldName, field.Int(), ruleParam, errors)
	case "in":
		return validateIn(fieldName, field.String(), ruleParam, errors)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownRule, ruleName)
	}
}

func validateString(fieldName string, field reflect.Value, ruleName string, ruleParam string, errors *ValidationErrors) error {
	switch ruleName {
	case "len":
		return validateLen(fieldName, field.String(), ruleParam, errors)
	case "in":
		return validateIn(fieldName, field.String(), ruleParam, errors)
	case "regexp":
		return validateRegexp(fieldName, field.String(), ruleParam, errors)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownRule, ruleName)
	}
}

func validateMax(fieldName string, fieldValue int64, param string, errors *ValidationErrors) error {
	vmax, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: max:%s", ErrInvalidParam, param)
	}

	if fieldValue > vmax {
		*errors = append(*errors, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("%w: %d > %d", ErrMaxExceeded, fieldValue, vmax),
		})
	}
	return nil
}

func validateMin(fieldName string, fieldValue int64, param string, errors *ValidationErrors) error {
	vmin, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: min:%s", ErrInvalidParam, param)
	}

	if fieldValue < vmin {
		*errors = append(*errors, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("%w: %d < %d", ErrMinNotReached, fieldValue, vmin),
		})
	}
	return nil
}

func validateIn(fieldName string, fieldValue string, param string, errors *ValidationErrors) error {
	in := strings.Split(param, ",")
	if len(in) == 0 {
		return fmt.Errorf("%w: in:%s", ErrInvalidParam, param)
	}

	if !slices.Contains(in, fieldValue) {
		*errors = append(*errors, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("%w: %s ", ErrNotInSet, fieldValue),
		})
	}
	return nil
}

func validateLen(fieldName string, fieldValue string, param string, errors *ValidationErrors) error {
	vlen, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: len:%s", ErrInvalidParam, param)
	}

	if utf8.RuneCountInString(fieldValue) != int(vlen) {
		*errors = append(*errors, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("%w: %s", ErrInvalidLength, fieldValue),
		})
	}
	return nil
}

func validateRegexp(fieldName string, fieldValue string, param string, errors *ValidationErrors) error {
	re, err := regexp.Compile(param)
	if err != nil {
		return fmt.Errorf("ошибка компиляции регулярного выражения regexp:%s %w", param, err)
	}

	if !re.MatchString(fieldValue) {
		*errors = append(*errors, ValidationError{
			Field: fieldName,
			Err:   fmt.Errorf("%w: %s", ErrNoMatchRegexp, fieldValue),
		})
	}
	return nil
}
