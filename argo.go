package argo

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type Error struct {
	Msg string
}

func NewError(msg string) *Error {
	return &Error{msg}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", argoTag, e.Msg)
}

const (
	argoTag                  string = "argo"
	shortAttribute           string = "short"
	longAttribute            string = "long"
	helpAttribute            string = "help"
	requiredAttribute        string = "required"
	errNotPointerToStruct    string = "argument must be a pointer to a struct"
	errAttributeMissingValue string = "attribute missing value"
	errUnknownAttribute      string = "unknown attribute"
	errMalformedAttribute    string = "malformed attribute"
	errAttributeInvalidValue string = "attribute value must be a flag name"
	errShortNotSingleChar    string = "short attribute must be a single character"
)

func Parse(args interface{}) error {
	value := reflect.ValueOf(args)
	valueType := value.Type()

	if valueType.Kind() != reflect.Ptr {
		return NewError(errNotPointerToStruct)
	}
	if valueType.Elem().Kind() != reflect.Struct {
		return NewError(errNotPointerToStruct)
	}

	elemType := value.Elem().Type()
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)

		tag := field.Tag.Get(argoTag)
		if tag == "" {
			continue
		}

		parsedAttributes, err := parseField(field, tag)
		if err != nil {
			return err
		}

		log.Printf("short: %s, long: %s, help: %s", parsedAttributes.short, parsedAttributes.long, parsedAttributes.help)
	}

	return nil
}

type fieldAttributes struct {
	short      string
	long       string
	help       string
	isRequired bool
}

func parseField(field reflect.StructField, tag string) (*fieldAttributes, error) {
	parsedAttributes := &fieldAttributes{}
	attributes := strings.Split(tag, ";")

	for _, attr := range attributes {
		if err := parseAttribute(field.Name, attr, parsedAttributes); err != nil {
			return nil, err
		}
	}
	return parsedAttributes, nil
}

func attributeToKeyValue(attribute string, fieldName string) (string, string, bool, error) {
	attrParts := strings.Split(attribute, "=")
	if len(attrParts) != 1 && len(attrParts) != 2 {
		return "", "", false, NewError(fmt.Sprintf("%s (%s)", errMalformedAttribute, attribute))
	}

	attrKey := attrParts[0]
	if len(attrParts) != 2 {
		return attrKey, strings.ToLower(fieldName), false, nil
	}

	return attrKey, attrParts[1], true, nil
}

func validateIdentifier(value string) error {
	matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", value)
	if err != nil || !matched {
		return NewError(errAttributeInvalidValue)
	}
	return nil
}

func parseAttribute(fieldName string, attribute string, parsedAttributes *fieldAttributes) error {
	attrKey, attrValue, hasValue, err := attributeToKeyValue(attribute, fieldName)
	if err != nil {
		return err
	}

	switch attrKey {
	case shortAttribute:
		if len(attrValue) != 1 {
			return NewError(fmt.Sprintf("%s (%s)", errShortNotSingleChar, attrValue))
		}
		if err := validateIdentifier(attrValue); err != nil {
			return err
		}

		parsedAttributes.short = attrValue[:1]
	case longAttribute:
		if err := validateIdentifier(attrValue); err != nil {
			return err
		}
		parsedAttributes.long = attrValue
	case helpAttribute:
		if !hasValue {
			return NewError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		} else {
			parsedAttributes.help = attrValue
		}
	case requiredAttribute:
		if hasValue {
			return NewError(fmt.Sprintf("%s (%s)", errMalformedAttribute, attrKey))
		}
		parsedAttributes.isRequired = true
	default:
		return NewError(fmt.Sprintf("%s (%s)", errUnknownAttribute, attrKey))
	}
	return nil
}
