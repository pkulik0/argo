package argo

import (
	"fmt"
	"log"
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

	errNotPointerToStruct    string = "argument must be a pointer to a struct"
	errAttributeMissingValue string = "attribute missing value"
	errUnknownAttribute      string = "unknown attribute"
	errMalformedAttribute    string = "malformed attribute"
	errAttributeInvalidValue string = "attribute has invalid value"
	errShortNotSingleChar    string = "short attribute value must be a single character"
	errUnsupportedType       string = "unsupported type"
	errSetterAlreadyExists   string = "setter already exists"
	errPositionalNotLast     string = "positional argument must be the last one"
	errPositionalConflict    string = "positional arguments cannot have short or long attributes"
	errDuplicateFlagName     string = "duplicate flag name"
	errUnknownFlagName       string = "unknown flag name"
	errUnexpectedArgument    string = "unexpected argument"
	errRequiredNotSet        string = "required argument not set"
	errPositionalNotSet      string = "positional argument not set"
	errFieldNotExported      string = "field is not exported"
	errCouldNotSet           string = "could not set value"
)

type setterFunc func(string, reflect.Value) error

type fieldSetterFunc func(string) error

var setters = map[reflect.Kind]setterFunc{
	reflect.String: func(s string, value reflect.Value) error {
		value.SetString(s)
		return nil
	},
	reflect.Bool: func(s string, value reflect.Value) error {
		boolValue, err := strconv.ParseBool(s)
		if err != nil {
			return newError(fmt.Sprintf("%s (%s)", errCouldNotSet, s))
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

func RegisterSetter(kind reflect.Kind, setter setterFunc) error {
	if _, ok := setters[kind]; ok {
		return newError(fmt.Sprintf("%s (%s)", errSetterAlreadyExists, kind))
	}
	setters[kind] = setter
	return nil
}

type field struct {
	short        string
	long         string
	isPositional bool
	isRequired   bool
	fromEnv      bool
	help         string
	defaultValue string
	setter       fieldSetterFunc
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

type registeredArgs struct {
	short      map[string]*field
	long       map[string]*field
	positional []*field
}

func newRegisteredArgs() *registeredArgs {
	return &registeredArgs{
		short:      make(map[string]*field),
		long:       make(map[string]*field),
		positional: make([]*field, 0),
	}
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

	registeredArgs, err := parseStruct(elem)
	if err != nil {
		return err
	}

	args := os.Args[1:]
	positionalIndex := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		log.Printf("arg: %s", arg)

		if strings.HasPrefix(arg, "-") {
			if positionalIndex != 0 {
				return newError(errPositionalConflict)
			}

			i++
			argName := arg[1:]
			argValue := args[i]
			var regArg *field

			if strings.HasPrefix(arg, "--") {
				argName = arg[2:]
				regArg = registeredArgs.long[argName]
			} else {
				regArg = registeredArgs.short[argName]
			}

			if regArg == nil {
				return newError(fmt.Sprintf("%s (%s)", errUnknownFlagName, argName))
			}
			if err := regArg.setter(argValue); err != nil {
				return err
			}
			regArg.wasSet = true
			continue
		}

		if len(registeredArgs.positional) == 0 || positionalIndex >= len(registeredArgs.positional) {
			return newError(fmt.Sprintf("%s (%s)", errUnexpectedArgument, arg))
		}

		regArg := registeredArgs.positional[positionalIndex]
		if err := regArg.setter(arg); err != nil {
			return err
		}
		regArg.wasSet = true
		positionalIndex++
	}

	for _, regArg := range registeredArgs.short {
		if regArg.wasSet {
			continue
		}

		if regArg.isPositional {
			return newError(errPositionalNotSet)
		}
		if regArg.isRequired {
			return newError(fmt.Sprintf("%s (%s)", errRequiredNotSet, regArg.short))
		}
		if regArg.fromEnv {
			envValue := os.Getenv(regArg.short)
			if envValue != "" {
				if err := regArg.setter(envValue); err != nil {
					return err
				}
				regArg.wasSet = true
			}
		}
		if regArg.defaultValue != "" {
			if err := regArg.setter(regArg.defaultValue); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseStruct(elem reflect.Value) (*registeredArgs, error) {
	registeredArgs := newRegisteredArgs()

	alreadyHasPositional := false
	for i := 0; i < elem.NumField(); i++ {
		value := elem.Field(i)
		structField := elem.Type().Field(i)

		if !structField.IsExported() {
			return nil, newError(fmt.Sprintf("%s (%s)", errFieldNotExported, structField.Name))
		}

		if structField.Tag.Get(argoTag) == "" {
			continue
		}

		parsedField, err := parseField(value, structField)
		if err != nil {
			return nil, err
		}
		log.Printf("%+v", parsedField)

		if parsedField.isPositional {
			registeredArgs.positional = append(registeredArgs.positional, parsedField)
			alreadyHasPositional = true
			continue
		}
		if alreadyHasPositional {
			return nil, newError(fmt.Sprintf("%s (%s)", errPositionalNotLast, structField.Name))
		}

		if parsedField.short != "" {
			if _, ok := registeredArgs.short[parsedField.short]; ok {
				return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, parsedField.short))
			}
			registeredArgs.short[parsedField.short] = parsedField
		}
		if parsedField.long != "" {
			if _, ok := registeredArgs.long[parsedField.long]; ok {
				return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, parsedField.long))
			}
			registeredArgs.long[parsedField.long] = parsedField
		}
	}
	return registeredArgs, nil
}

func parseField(fieldValue reflect.Value, structField reflect.StructField) (*field, error) {
	parsedField := &field{}
	attributes := strings.Split(structField.Tag.Get(argoTag), attributeSeparator)

	for _, attr := range attributes {
		if err := parseAttribute(structField.Name, attr, parsedField); err != nil {
			return nil, err
		}
	}

	if parsedField.short == "" && parsedField.long == "" && !parsedField.isPositional {
		fieldName := strings.ToLower(structField.Name)
		parsedField.short = fieldName[:1]
		parsedField.long = fieldName
	}

	setter, ok := setters[fieldValue.Kind()]
	if !ok {
		return nil, newError(fmt.Sprintf("%s (%s)", errUnsupportedType, fieldValue.Kind()))
	}

	parsedField.setter = func(value string) error {
		return setter(value, fieldValue)
	}

	return parsedField, nil
}

func attributeToKeyValue(attribute string) (string, string, error) {
	attrParts := strings.Split(attribute, attributeValueSeparator)
	if len(attrParts) != 1 && len(attrParts) != 2 {
		return "", "", newError(fmt.Sprintf("%s (%s)", errMalformedAttribute, attribute))
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

func parseAttribute(fieldName string, attribute string, parsedAttributes *field) error {
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
	case positionalAttribute:
		return parseAttributeBool(attrValue, &parsedAttributes.isPositional)
	case requiredAttribute:
		return parseAttributeBool(attrValue, &parsedAttributes.isRequired)
	case envAttribute:
		return parseAttributeBool(attrValue, &parsedAttributes.fromEnv)
	case helpAttribute:
		if attrValue == "" {
			return newError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		parsedAttributes.help = attrValue
	case defaultAttribute:
		if attrValue == "" {
			return newError(fmt.Sprintf("%s (%s)", errAttributeMissingValue, attrKey))
		}
		parsedAttributes.defaultValue = attrValue
	default:
		return newError(fmt.Sprintf("%s (%s)", errUnknownAttribute, attrKey))
	}
	return nil
}

func setterInt(value string, out reflect.Value, bitSize int) error {
	intValue, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return newError(fmt.Sprintf("%s (%s)", errCouldNotSet, value))
	}
	out.SetInt(intValue)
	return nil
}

func setterUint(value string, out reflect.Value, bitSize int) error {
	uintValue, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return newError(fmt.Sprintf("%s (%s)", errCouldNotSet, value))
	}
	out.SetUint(uintValue)
	return nil
}

func setterFloat(value string, out reflect.Value, bitSize int) error {
	floatValue, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return newError(fmt.Sprintf("%s (%s)", errCouldNotSet, value))
	}
	out.SetFloat(floatValue)
	return nil
}
