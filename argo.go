package argo

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	argoTag string = "argo"

	shortAttribute      string = "short"
	longAttribute       string = "long"
	positionalAttribute string = "positional"
	helpAttribute       string = "help"
	requiredAttribute   string = "required"
	envAttribute        string = "env"
	defaultAttribute    string = "default"

	attributeSeparator      string = ","
	attributeValueSeparator string = "="

	errNotPointerToStruct       string = "argument must be a pointer to a struct"
	errAttributeMissingValue    string = "attribute missing value"
	errUnknownAttribute         string = "unknown attribute"
	errMalformedAttribute       string = "malformed attribute"
	errAttributeInvalidValue    string = "attribute has invalid value"
	errShortNotSingleChar       string = "short attribute value must be a single character"
	errUnsupportedType          string = "unsupported type"
	errSetterAlreadyExists      string = "setter already exists"
	errPositionalNotAtEnd       string = "positional arguments must be at the end"
	errPositionalConflict       string = "positional arguments cannot have short or long attributes"
	errPositionalDefaultNotLast string = "positional arguments can have a default value only if no arguments without one follow"
	errDuplicateFlagName        string = "duplicate flag name, consider changing the short or long attribute"
	errUnknownFlagName          string = "unknown flag name"
	errUnexpectedArgument       string = "unexpected argument"
	errRequiredNotSet           string = "required argument not set"
	errPositionalNotSet         string = "positional argument not set"
	errFieldNotExported         string = "field must be exported"
	errCouldNotSet              string = "could not set value"
)

type setterFunc func(string, reflect.Value) error

var setters = map[reflect.Kind]setterFunc{
	reflect.String: func(s string, value reflect.Value) error {
		value.SetString(s)
		return nil
	},
	reflect.Bool: func(s string, value reflect.Value) error {
		boolValue, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		value.SetBool(boolValue)
		return nil
	},
	reflect.Int: func(s string, value reflect.Value) error {
		return setterInt(s, value, 0)
	},
	reflect.Int8: func(s string, value reflect.Value) error {
		return setterInt(s, value, 8)
	},
	reflect.Int16: func(s string, value reflect.Value) error {
		return setterInt(s, value, 16)
	},
	reflect.Int32: func(s string, value reflect.Value) error {
		return setterInt(s, value, 32)
	},
	reflect.Int64: func(s string, value reflect.Value) error {
		return setterInt(s, value, 64)
	},
	reflect.Uint: func(s string, value reflect.Value) error {
		return setterUint(s, value, 0)
	},
	reflect.Uint8: func(s string, value reflect.Value) error {
		return setterUint(s, value, 8)
	},
	reflect.Uint16: func(s string, value reflect.Value) error {
		return setterUint(s, value, 16)
	},
	reflect.Uint32: func(s string, value reflect.Value) error {
		return setterUint(s, value, 32)
	},
	reflect.Uint64: func(s string, value reflect.Value) error {
		return setterUint(s, value, 64)
	},
	reflect.Float32: func(s string, value reflect.Value) error {
		return setterFloat(s, value, 32)
	},
	reflect.Float64: func(s string, value reflect.Value) error {
		return setterFloat(s, value, 64)
	},
	reflect.Interface: func(s string, value reflect.Value) error {
		value.Set(reflect.ValueOf(s))
		return nil
	},
}

func RegisterSetter(t interface{}, setter setterFunc) error {
	kind := reflect.TypeOf(t).Kind()
	if _, ok := setters[kind]; ok {
		return newError(fmt.Sprintf("%s (%s)", errSetterAlreadyExists, kind))
	}
	setters[kind] = setter
	return nil
}

type arg struct {
	name         string
	short        string
	long         string
	env          string
	isPositional bool
	isRequired   bool
	help         string
	defaultValue string
	setter       func(string) error
	wasSet       bool
}

type argoError struct {
	Msg string
}

func (e *argoError) Error() string {
	return fmt.Sprintf("%s: %s", argoTag, e.Msg)
}

func newError(msg string) *argoError {
	return &argoError{msg}
}

type argsRegistry struct {
	short      map[string]*arg
	long       map[string]*arg
	positional []*arg
}

func newArgsRegistry() *argsRegistry {
	return &argsRegistry{
		short:      make(map[string]*arg),
		long:       make(map[string]*arg),
		positional: make([]*arg, 0),
	}
}

func (r *argsRegistry) Range() <-chan *arg {
	ch := make(chan *arg)
	go func() {
		for _, argument := range r.short {
			ch <- argument
		}
		for _, argument := range r.long {
			ch <- argument
		}
		for _, argument := range r.positional {
			ch <- argument
		}
		close(ch)
	}()
	return ch
}

func Parse(outputStruct interface{}) error {
	outputValue := reflect.ValueOf(outputStruct)

	if outputValue.Kind() != reflect.Ptr || outputValue.IsNil() {
		return newError(errNotPointerToStruct)
	}

	elem := outputValue.Elem()
	if elem.Kind() != reflect.Struct {
		return newError(errNotPointerToStruct)
	}

	argumentsRegistry, err := parseStruct(elem)
	if err != nil {
		return err
	}

	args := os.Args[1:]
	positionalIndex := 0
	for i := 0; i < len(args); i++ {
		argText := args[i]

		if strings.HasPrefix(argText, "-") {
			if positionalIndex != 0 {
				return newError(errPositionalConflict)
			}

			i++
			argName := argText[1:]
			argValue := args[i]
			var argument *arg

			if strings.HasPrefix(argText, "--") {
				argName = argText[2:]
				argument = argumentsRegistry.long[argName]
			} else {
				argument = argumentsRegistry.short[argName]
			}

			if argument == nil {
				return newError(fmt.Sprintf("%s (%s)", errUnknownFlagName, argName))
			}
			if err := argument.setter(argValue); err != nil {
				return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argText))
			}

			continue
		}

		if len(argumentsRegistry.positional) == 0 || positionalIndex >= len(argumentsRegistry.positional) {
			return newError(fmt.Sprintf("%s (%s)", errUnexpectedArgument, argText))
		}

		argument := argumentsRegistry.positional[positionalIndex]
		if err := argument.setter(argText); err != nil {
			return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argText))
		}
		positionalIndex++
	}

	alreadyHasPositional := false
	for argument := range argumentsRegistry.Range() {
		if argument.wasSet {
			continue
		}

		if argument.isPositional {
			alreadyHasPositional = true
			if argument.defaultValue != "" {
				if err := argument.setter(argument.defaultValue); err != nil {
					return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argument.defaultValue))
				}
				continue
			}
			return newError(errPositionalNotSet)
		}
		if alreadyHasPositional {
			return newError(errPositionalNotAtEnd)
		}

		if argument.env != "" {
			envValue := os.Getenv(argument.env)
			if envValue != "" {
				if err := argument.setter(envValue); err != nil {
					return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, envValue))
				}
			}
		}
		if argument.defaultValue != "" {
			if err := argument.setter(argument.defaultValue); err != nil {
				return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argument.defaultValue))
			}
		}
		if argument.isRequired {
			return newError(fmt.Sprintf("%s (%s)", errRequiredNotSet, argument.short))
		}
	}

	return nil
}

func parseStruct(elem reflect.Value) (*argsRegistry, error) {
	registeredArgs := newArgsRegistry()

	hasPositional := false
	hasDefaultedPositional := false
	for i := 0; i < elem.NumField(); i++ {
		value := elem.Field(i)
		structField := elem.Type().Field(i)

		if !structField.IsExported() {
			return nil, newError(fmt.Sprintf("%s (%s)", errFieldNotExported, structField.Name))
		}

		if structField.Tag.Get(argoTag) == "" {
			continue
		}

		argument, err := parseArgument(value, structField)
		if err != nil {
			return nil, err
		}

		if argument.isPositional {
			if hasDefaultedPositional {
				return nil, newError(fmt.Sprintf("%s (%s)", errPositionalDefaultNotLast, structField.Name))
			}

			registeredArgs.positional = append(registeredArgs.positional, argument)

			hasPositional = true
			if argument.defaultValue != "" {
				hasDefaultedPositional = true
			}
		} else {
			if hasPositional {
				return nil, newError(fmt.Sprintf("%s (%s)", errPositionalNotAtEnd, structField.Name))
			}

			if argument.short != "" {
				if _, ok := registeredArgs.short[argument.short]; ok {
					return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, argument.short))
				}
				registeredArgs.short[argument.short] = argument
			}
			if argument.long != "" {
				if _, ok := registeredArgs.long[argument.long]; ok {
					return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, argument.long))
				}
				registeredArgs.long[argument.long] = argument
			}
		}
	}
	return registeredArgs, nil
}

func parseArgument(fieldValue reflect.Value, structField reflect.StructField) (*arg, error) {
	argument := &arg{
		name: structField.Name,
	}
	attributes := strings.Split(structField.Tag.Get(argoTag), attributeSeparator)

	for _, attr := range attributes {
		if err := parseAttribute(structField.Name, attr, argument); err != nil {
			return nil, err
		}
	}

	if argument.short == "" && argument.long == "" && !argument.isPositional {
		fieldName := strings.ToLower(structField.Name)
		argument.short = fieldName[:1]
		argument.long = fieldName
	}

	setter, ok := setters[fieldValue.Kind()]
	if !ok {
		return nil, newError(fmt.Sprintf("%s (%s)", errUnsupportedType, fieldValue.Kind()))
	}

	argument.setter = func(value string) error {
		argument.wasSet = true
		return setter(value, fieldValue)
	}

	return argument, nil
}

func attributeToKeyValue(attribute string) (string, string, error) {
	attrParts := strings.Split(attribute, attributeValueSeparator)
	if len(attrParts) != 1 && len(attrParts) != 2 {
		return "", "", newError(fmt.Sprintf("%s (%s)", errMalformedAttribute, attribute))
	}

	attrKey := attrParts[0]
	if len(attrParts) == 1 {
		return attrKey, "", nil
	}
	return attrKey, attrParts[1], nil
}

func validateIdentifier(value string) error {
	matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]*$", value)
	if err != nil || !matched {
		return newError(errAttributeInvalidValue)
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
		return newError(fmt.Sprintf("%s (%s)", errAttributeInvalidValue, value))
	}
	*out = boolValue

	return nil
}

func parseAttributeIdentifier(value string, defaultValue string, out *string) error {
	if value == "" {
		*out = defaultValue
		return nil
	}

	if err := validateIdentifier(value); err != nil {
		return err
	}

	*out = value
	return nil
}

func parseAttribute(fieldName string, attribute string, argument *arg) error {
	if attribute == "" {
		return newError(errMalformedAttribute)
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
				return newError(fmt.Sprintf("%s (%s)", errShortNotSingleChar, attrValue))
			}
			if err := validateIdentifier(attrValue); err != nil {
				return err
			}
		}
		argument.short = attrValue[:1]
	case longAttribute:
		return parseAttributeIdentifier(attrValue, strings.ToLower(fieldName), &argument.long)
	case positionalAttribute:
		return parseAttributeBool(attrValue, &argument.isPositional)
	case requiredAttribute:
		return parseAttributeBool(attrValue, &argument.isRequired)
	case envAttribute:
		return parseAttributeIdentifier(attrValue, strings.ToUpper(fieldName), &argument.env)
	case helpAttribute:
		if attrValue == "" {
			return newError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		argument.help = attrValue
	case defaultAttribute:
		if attrValue == "" {
			return newError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		argument.defaultValue = attrValue
	default:
		return newError(fmt.Sprintf("%s (%s)", errUnknownAttribute, attrKey))
	}
	return nil
}

func setterInt(value string, out reflect.Value, bitSize int) error {
	intValue, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return err
	}
	out.SetInt(intValue)
	return nil
}

func setterUint(value string, out reflect.Value, bitSize int) error {
	uintValue, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return err
	}
	out.SetUint(uintValue)
	return nil
}

func setterFloat(value string, out reflect.Value, bitSize int) error {
	floatValue, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}
	out.SetFloat(floatValue)
	return nil
}
