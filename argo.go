package argo

import (
	"errors"
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
	errPositionalDefaultNotLast string = "positional arguments can have a default value only if no arguments without one follow"
	errDuplicateFlagName        string = "duplicate flag name, consider changing the short or long attribute"
	errUnknownArgumentName      string = "unknown argument name"
	errUnexpectedArgument       string = "unexpected argument"
	errRequiredNotSet           string = "required argument not set"
	errPositionalNotSet         string = "positional argument not set"
	errFieldNotExported         string = "field must be exported"
	errCouldNotSet              string = "could not set value"
)

type arg struct {
	name         string
	short        string
	long         string
	env          string
	isPositional bool
	isRequired   bool
	isFlag       bool
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
	env        map[string]*arg
	positional []*arg
}

func (r *argsRegistry) asRange() <-chan *arg {
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
		for _, argument := range r.env {
			ch <- argument
		}
		close(ch)
	}()
	return ch
}

func (r *argsRegistry) deduplicated() map[*arg]struct{} {
	dedup := make(map[*arg]struct{})
	for argument := range r.asRange() {
		dedup[argument] = struct{}{}
	}
	return dedup
}

func interfaceToArgsRegistry(input interface{}) (*argsRegistry, error) {
	outputValue := reflect.ValueOf(input)

	if outputValue.Kind() != reflect.Ptr || outputValue.IsNil() {
		return nil, newError(errNotPointerToStruct)
	}

	elem := outputValue.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, newError(errNotPointerToStruct)
	}

	return newArgsRegistry(elem)
}

func Parse(input interface{}) error {
	argumentsRegistry, err := interfaceToArgsRegistry(input)
	if err != nil {
		return err
	}

	if err = argumentsRegistry.parseInput(); err != nil {
		return err
	}

	return validateArgsRegistry(argumentsRegistry)
}

func PrintHelp(input interface{}) error {
	argumentsRegistry, err := interfaceToArgsRegistry(input)
	if err != nil {
		return err
	}
	return argumentsRegistry.printHelp()
}

func (r *argsRegistry) printHelp() error {
	flags := make([]string, 0)
	positionals := make([]string, 0)
	envs := make([]string, 0)

	for argument := range r.deduplicated() {
		if argument.isPositional {
			positionals = append(positionals, fmt.Sprintf("<%s>", argument.name))
			continue
		}

		flag := ""
		if argument.short != "" {
			flag += fmt.Sprintf("-%s", argument.short)
		}

		if argument.long != "" {
			if flag != "" {
				flag += ", "
			}
			flag += fmt.Sprintf("--%s", argument.long)
		}

		hasFlag := flag != ""
		if argument.env != "" {
			if hasFlag {
				flag += " "
			}
			flag += fmt.Sprintf("[ENV: %s]", argument.env)
		}

		if argument.help != "" {
			flag += fmt.Sprintf(" - %s", argument.help)
		}

		if argument.defaultValue != "" {
			flag += fmt.Sprintf(" (default: %s)", argument.defaultValue)
		}

		if argument.isRequired {
			flag += " (REQUIRED)"
		}

		if hasFlag {
			flags = append(flags, flag)
		} else {
			envs = append(envs, flag)
		}
	}

	output := fmt.Sprintf("Usage: ./%s [flags] %s\n", os.Args[0], strings.Join(positionals, " "))

	flags = append(flags, " -h, --help - Print this help message")
	output += "\nFlags:\n"
	for _, flag := range flags {
		output += fmt.Sprintf("  %s\n", flag)
	}

	if len(envs) > 0 {
		output += "\nEnvironment variables:\n"
		for _, env := range envs {
			output += fmt.Sprintf("  %s\n", env)
		}
	}

	return errors.New(output)
}

func (r *argsRegistry) parseInput() error {
	args := os.Args[1:]
	positionalIndex := 0
	explicitPositional := false
	for i := 0; i < len(args); i++ {
		argText := args[i]

		if argText == "--" {
			explicitPositional = true
			continue
		}

		if argText == "-h" || argText == "--help" {
			return r.printHelp()
		}

		if strings.HasPrefix(argText, "-") && !explicitPositional {
			if positionalIndex != 0 {
				return newError(errPositionalNotAtEnd)
			}

			argName := argText[1:]
			var argument *arg

			if strings.HasPrefix(argText, "--") {
				argName = argText[2:]
				argument = r.long[argName]
			} else {
				argument = r.short[argName]
			}

			if argument == nil {
				return newError(fmt.Sprintf("%s (%s)", errUnknownArgumentName, argName))
			}

			if !argument.isFlag {
				i++
				err := argument.setter(args[i])
				if err != nil {
					return newError(fmt.Sprintf("%s (%s = %s): %s", errCouldNotSet, argument.name, args[i], err))
				}
			} else {
				_ = argument.setter("true")
			}

			continue
		}

		if len(r.positional) == 0 || positionalIndex >= len(r.positional) {
			return newError(fmt.Sprintf("%s (%s)", errUnexpectedArgument, argText))
		}

		argument := r.positional[positionalIndex]
		if err := argument.setter(argText); err != nil {
			return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argText))
		}
		positionalIndex++
	}
	return nil
}

func validateArgsRegistry(argumentsRegistry *argsRegistry) error {
	for argument := range argumentsRegistry.deduplicated() {
		if argument.wasSet {
			continue
		}

		if argument.isPositional {
			if argument.defaultValue != "" {
				if err := argument.setter(argument.defaultValue); err != nil {
					return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argument.defaultValue))
				}
				continue
			}
			return newError(errPositionalNotSet)
		}

		if argument.env != "" {
			envValue := os.Getenv(argument.env)
			if envValue != "" {
				if err := argument.setter(envValue); err != nil {
					return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, envValue))
				}
				continue
			}
		}

		if argument.defaultValue != "" {
			if err := argument.setter(argument.defaultValue); err != nil {
				return newError(fmt.Sprintf("%s (%s = %s)", errCouldNotSet, argument.name, argument.defaultValue))
			}
			continue
		}

		if argument.isRequired {
			return newError(fmt.Sprintf("%s (%s)", errRequiredNotSet, argument.short))
		}
	}
	return nil
}

func newArgsRegistry(elem reflect.Value) (*argsRegistry, error) {
	registeredArgs := &argsRegistry{
		short:      make(map[string]*arg),
		long:       make(map[string]*arg),
		positional: make([]*arg, 0),
		env:        make(map[string]*arg),
	}

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

			if argument.defaultValue != "" {
				hasDefaultedPositional = true
			}
			continue
		}

		if argument.env != "" {
			if _, ok := registeredArgs.env[argument.env]; ok {
				return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, argument.env))
			}
			registeredArgs.env[argument.env] = argument
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

	if argument.short == "" && argument.long == "" && !argument.isPositional && argument.env == "" {
		fieldName := strings.ToLower(structField.Name)
		argument.short = fieldName[:1]
		argument.long = fieldName
	}

	if argument.short == "h" || argument.long == "help" {
		return nil, newError(fmt.Sprintf("%s (%s)", errDuplicateFlagName, argument.short))
	}

	kind := structField.Type.Kind()
	isPtr := kind == reflect.Ptr
	if isPtr {
		kind = structField.Type.Elem().Kind()
	}

	setter, ok := setters[kind]
	if !ok {
		return nil, newError(fmt.Sprintf("%s (%s)", errUnsupportedType, kind))
	}

	argument.setter = func(value string) error {
		if isPtr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(structField.Type.Elem()))
			}
			fieldValue = fieldValue.Elem()
		}

		argument.wasSet = true
		return setter(value, fieldValue)
	}

	if kind == reflect.Bool {
		argument.isFlag = true
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
		return newError(fmt.Sprintf("%s (%s)", errAttributeInvalidValue, value))
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
	attrKey, attrValue, err := attributeToKeyValue(attribute)
	if err != nil {
		return err
	}

	switch attrKey {
	case shortAttribute:
		if attrValue == "" {
			attrValue = strings.ToLower(fieldName)
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
