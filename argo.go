package argo

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
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
	argoTag string = "argo"

	attributeSeparator      string = ","
	attributeValueSeparator string = "="
	shortAttribute          string = "short"
	longAttribute           string = "long"
	helpAttribute           string = "help"
	requiredAttribute       string = "required"
	envAttribute            string = "env"
	defaultAttribute        string = "default"

	errNotPointerToStruct    string = "argument must be a pointer to a struct"
	errAttributeMissingValue string = "attribute missing value"
	errUnknownAttribute      string = "unknown attribute"
	errMalformedAttribute    string = "malformed attribute"
	errAttributeInvalidValue string = "attribute has invalid value"
	errShortNotSingleChar    string = "short attribute value must be a single character"
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

		attributes, err := parseField(field.Name, tag)
		if err != nil {
			return err
		}

		log.Printf("short: %s, long: %s, help: %s", attributes.short, attributes.long, attributes.help)
	}

	return nil
}

type fieldAttributes struct {
	short        string
	long         string
	help         string
	isRequired   bool
	fromEnv      bool
	defaultValue string
}

func parseField(fieldName string, tag string) (*fieldAttributes, error) {
	parsedAttributes := &fieldAttributes{}
	attributes := strings.Split(tag, attributeSeparator)

	for _, attr := range attributes {
		if err := parseAttribute(fieldName, attr, parsedAttributes); err != nil {
			return nil, err
		}
	}
	return parsedAttributes, nil
}

func attributeToKeyValue(attribute string) (string, string, error) {
	attrParts := strings.Split(attribute, attributeValueSeparator)
	if len(attrParts) != 1 && len(attrParts) != 2 {
		return "", "", NewError(fmt.Sprintf("%s (%s)", errMalformedAttribute, attribute))
	}

	attrKey := attrParts[0]
	if len(attrParts) != 2 {
		return attrKey, "", nil
	}

	return attrKey, attrParts[1], nil
}

func validateIdentifier(value string) error {
	matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", value)
	if err != nil || !matched {
		return NewError(errAttributeInvalidValue)
	}
	return nil
}

func parseAttributeBool(value string, out *bool) error {
	if value == "" {
		*out = true
		return nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return NewError(fmt.Sprintf("%s (%s)", errAttributeInvalidValue, value))
	}
	*out = boolValue

	return nil
}

func parseAttribute(fieldName string, attribute string, parsedAttributes *fieldAttributes) error {
	if attribute == "" {
		return NewError(errMalformedAttribute)
	}

	attrKey, attrValue, err := attributeToKeyValue(attribute)
	if err != nil {
		return err
	}

	switch attrKey {
	case shortAttribute:
		if attrValue == "" {
			attrValue = fieldName[:1]
		} else {
			if len(attrValue) != 1 {
				return NewError(fmt.Sprintf("%s (%s)", errShortNotSingleChar, attrValue))
			}
			if err := validateIdentifier(attrValue); err != nil {
				return err
			}
		}
		parsedAttributes.short = attrValue[:1]
	case longAttribute:
		if attrValue == "" {
			attrValue = fieldName
		} else {
			if err := validateIdentifier(attrValue); err != nil {
				return err
			}
		}
		parsedAttributes.long = attrValue
	case helpAttribute:
		if attrValue == "" {
			return NewError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		parsedAttributes.help = attrValue
	case requiredAttribute:
		return parseAttributeBool(attrValue, &parsedAttributes.isRequired)
	case envAttribute:
		return parseAttributeBool(attrValue, &parsedAttributes.fromEnv)
	case defaultAttribute:
		if attrValue == "" {
			return NewError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		parsedAttributes.defaultValue = attrValue
	default:
		return NewError(fmt.Sprintf("%s (%s)", errUnknownAttribute, attrKey))
	}
	return nil
}
