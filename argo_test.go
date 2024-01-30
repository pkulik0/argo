package argo

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestParseAttributesFlags(t *testing.T) {
	attribs := &arg{}
	if err := parseAttribute("name", "short=s", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "s" {
		t.Fatalf("expected 's', got '%s'", attribs.short)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "short", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "n" {
		t.Fatalf("expected 'n', got '%s'", attribs.short)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "long", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "name" {
		t.Fatalf("expected 'name', got '%s'", attribs.long)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "long=longname", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "longname" {
		t.Fatalf("expected 'longname', got '%s'", attribs.long)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "positional", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.isPositional {
		t.Fatal("expected 'isPositional' to be true")
	}
}

func TestParseAttributesBool(t *testing.T) {
	attribs := &arg{}
	if err := parseAttribute("name", "required", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.isRequired {
		t.Fatal("expected 'isRequired' to be true")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "required=false", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.isRequired {
		t.Fatal("expected 'isRequired' to be false")
	}
}

func TestParseAttributesString(t *testing.T) {
	attribs := &arg{}
	if err := parseAttribute("name", "help=help text", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.help != "help text" {
		t.Fatalf("expected 'help text', got '%s'", attribs.help)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "default=123,abc", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.defaultValue != "123,abc" {
		t.Fatalf("expected '123,abc', got '%s'", attribs.defaultValue)
	}

	attribs = &arg{}
	if err := parseAttribute("name", "env", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.env != "NAME" {
		t.Fatalf("expected 'env' to be 'NAME', got '%s'", attribs.env)
	}

	attribs = &arg{}
	if err := parseAttribute("abc_123", "env=abc_123", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.env != "abc_123" {
		t.Fatalf("expected 'env' to be 'abc_123', got '%s'", attribs.env)
	}
}

func TestParseAttributesInvalid(t *testing.T) {
	attribs := &arg{}
	if err := parseAttribute("name", "help", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "short=3", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "short=ab_uu", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "long= eoeeo", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "required=something", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "env=321", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "default", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "default=", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "unknown", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "positional=123", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "short=abc=123;long", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "short;;", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &arg{}
	if err := parseAttribute("name", "long=abc; ;", attribs); err == nil {
		t.Fatal("expected error")
	}
}

type parseFieldStruct struct {
	name string `argo:"short=s,long=longname,help=help text,required,env,default=123abc"`
	age  int    `argo:"short,long=abc123,required=False,default=default value"`
}

func TestParseField(t *testing.T) {
	data := reflect.ValueOf(&parseFieldStruct{})
	fieldValue := data.Elem().Field(0)
	structField := data.Elem().Type().Field(0)

	attribs, err := parseArgument(fieldValue, structField)
	if err != nil {
		t.Fatal(err)
	}
	if attribs.short != "s" {
		t.Fatalf("expected 's', got '%s'", attribs.short)
	}
	if attribs.long != "longname" {
		t.Fatalf("expected 'longname', got '%s'", attribs.long)
	}
	if attribs.help != "help text" {
		t.Fatalf("expected 'help text', got '%s'", attribs.help)
	}
	if !attribs.isRequired {
		t.Fatal("expected 'isRequired' to be true")
	}
	if attribs.env != "NAME" {
		t.Fatalf("expected 'env' to be 'NAME', got '%s'", attribs.env)
	}
	if attribs.defaultValue != "123abc" {
		t.Fatalf("expected '123abc', got '%s'", attribs.defaultValue)
	}
	if attribs.setter == nil {
		t.Fatal("expected 'setter' to be set")
	}

	fieldValue = data.Elem().Field(1)
	structField = data.Elem().Type().Field(1)

	attribs, err = parseArgument(fieldValue, structField)
	if err != nil {
		t.Fatal(err)
	}
	if attribs.short != "a" {
		t.Fatalf("expected 'a', got '%s'", attribs.short)
	}
	if attribs.long != "abc123" {
		t.Fatalf("expected 'abc123', got '%s'", attribs.long)
	}
	if attribs.help != "" {
		t.Fatalf("expected '', got '%s'", attribs.help)
	}
	if attribs.isRequired {
		t.Fatal("expected 'isRequired' to be false")
	}
	if attribs.env != "" {
		t.Fatalf("expected env to be '', got '%s'", attribs.env)
	}
	if attribs.defaultValue != "default value" {
		t.Fatalf("expected 'default value', got '%s'", attribs.defaultValue)
	}
	if attribs.setter == nil {
		t.Fatal("expected 'setter' to be set")
	}
}

type argsEmptyTag struct {
	Host string `argo:""`
}

func TestEmpty(t *testing.T) {
	os.Args = []string{"test"}
	var args struct{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}

	os.Args = []string{"test", "--host", "localhost"}
	args2 := argsEmptyTag{}
	if err := Parse(&args2); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotPointerOrStruct(t *testing.T) {
	os.Args = []string{"test"}
	if err := Parse(struct{}{}); err == nil {
		t.Fatal("expected error")
	}

	test := 123
	if err := Parse(&test); err == nil {
		t.Fatal("expected error")
	}
}

type argsSimple struct {
	Address string `argo:"short=a,long=addr,help=Address to connect to"`
}

func TestSimple(t *testing.T) {
	os.Args = []string{"test"}
	args := argsSimple{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}

	os.Args = []string{"test", "-a", "localhost"}
	args = argsSimple{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Address != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Address)
	}

	os.Args = []string{"test", "--addr", "127.0.0.1"}
	args = argsSimple{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Address != "127.0.0.1" {
		t.Fatalf("expected '127.0.0.1', got '%s'", args.Address)
	}

}

type argsPositional struct {
	Host string `argo:"positional"`
	Port int    `argo:"positional"`
}

func TestPositional(t *testing.T) {
	os.Args = []string{"test", "localhost", "1234"}
	args := argsPositional{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Host)
	}
	if args.Port != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.Port)
	}
}

type argsPositionalDefault struct {
	Host string `argo:"positional"`
	Port int    `argo:"positional,default=1234"`
}

func TestPositionalDefault(t *testing.T) {
	os.Args = []string{"test", "localhost"}
	args := argsPositionalDefault{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Host)
	}
	if args.Port != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.Port)
	}
}

type argsPositionalDefaultInvalid struct {
	Host  string `argo:"positional"`
	Port  int    `argo:"positional,default=1234"`
	Speed int    `argo:"positional"`
}

type argsPositionalDefaultInvalid2 struct {
	Host string `argo:"long"`
	Port int    `argo:"positional"`
}

func TestPositionalDefaultInvalid(t *testing.T) {
	os.Args = []string{"test", "localhost"}
	args := argsPositionalDefaultInvalid{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	args2 := argsPositionalDefaultInvalid2{}
	if err := Parse(&args2); err == nil {
		t.Fatal("expected error")
	}
}

type argsTestLong struct {
	Host string `argo:"long=hostname"`
	Port int    `argo:"long=long"`
}

func TestLong(t *testing.T) {
	run := func() {
		args := argsTestLong{}
		if err := Parse(&args); err != nil {
			t.Fatal(err)
		}
		if args.Host != "localhost" {
			t.Fatalf("expected 'localhost', got '%s'", args.Host)
		}
		if args.Port != 1234 {
			t.Fatalf("expected '1234', got '%d'", args.Port)
		}
	}
	os.Args = []string{"test", "--hostname", "localhost", "--long", "1234"}
	run()
	os.Args = []string{"test", "--long", "1234", "--hostname", "localhost"}
	run()
}

type argsTestShort struct {
	Host string `argo:"short=h"`
	Port int    `argo:"short=p"`
}

func TestShort(t *testing.T) {
	run := func() {
		args := argsTestShort{}
		if err := Parse(&args); err != nil {
			t.Fatal(err)
		}
		if args.Host != "localhost" {
			t.Fatalf("expected 'localhost', got '%s'", args.Host)
		}
		if args.Port != 1234 {
			t.Fatalf("expected '1234', got '%d'", args.Port)
		}
	}
	os.Args = []string{"test", "-h", "localhost", "-p", "1234"}
	run()
	os.Args = []string{"test", "-p", "1234", "-h", "localhost"}
	run()
}

type argsTestRequired struct {
	Host string `argo:"short=h"`
	Port int    `argo:"short=p,required"`
}

func TestRequired(t *testing.T) {
	os.Args = []string{"test", "-h", "localhost", "-p", "1234"}
	args := argsTestRequired{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Host)
	}
	if args.Port != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.Port)
	}

	os.Args = []string{"test", "-h", "localhost"}
	args = argsTestRequired{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsTestRequired2 struct {
	Host     string `argo:"short=h"`
	Port     int    `argo:"short=p,required"`
	Username string `argo:"short=u,required"`
}

func TestRequired2(t *testing.T) {
	os.Args = []string{"test", "-h", "localhost", "-p", "1234", "-u", "user"}
	args := argsTestRequired2{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Host)
	}
	if args.Port != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.Port)
	}
	if args.Username != "user" {
		t.Fatalf("expected 'user', got '%s'", args.Username)
	}

	os.Args = []string{"test", "-h", "localhost", "-p", "1234"}
	args = argsTestRequired2{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	os.Args = []string{"test", "-h", "localhost", "-u", "user"}
	args = argsTestRequired2{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsTestEnv struct {
	Host string `argo:"long=host,env=HOST"`
	Port int    `argo:"env=PORT"`
}

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
}

func TestEnv(t *testing.T) {
	setEnv(t, "HOST", "localhost")
	setEnv(t, "PORT", "1234")

	os.Args = []string{"test"}
	args := argsTestEnv{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "localhost" {
		t.Fatalf("expected 'localhost', got '%s'", args.Host)
	}
	if args.Port != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.Port)
	}

	os.Args = []string{"test", "--host", "127.0.0.1"}
	args = argsTestEnv{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Host != "127.0.0.1" {
		t.Fatalf("expected '127.0.0.1', got '%s'", args.Host)
	}
}

func TestArgoError(t *testing.T) {
	os.Args = []string{"test", "--host", "localhost", "--port", "abc"}
	args := argsTestEnv{}
	err := Parse(&args)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(err.Error(), "argo: ") {
		t.Fatalf("expected 'argo' prefix, got '%s'", err.Error())
	}
}

type argsInts struct {
	A int   `argo:"short=a"`
	B int8  `argo:"long=b"`
	C int16 `argo:"env=C_VALUE"`
	D int32 `argo:"short,default=1234"`
	E int64 `argo:"positional"`
}

func TestAllInts(t *testing.T) {
	if err := os.Setenv("C_VALUE", "123"); err != nil {
		t.Fatalf("failed to set env: %s", err)
	}
	os.Args = []string{"test", "-a", "-5", "--b", "127", "--", "-100"}

	args := argsInts{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A != -5 {
		t.Fatalf("expected '-5', got '%d'", args.A)
	}
	if args.B != 127 {
		t.Fatalf("expected '127', got '%d'", args.B)
	}
	if args.C != 123 {
		t.Fatalf("expected '123', got '%d'", args.C)
	}
	if args.D != 1234 {
		t.Fatalf("expected '1234', got '%d'", args.D)
	}
	if args.E != -100 {
		t.Fatalf("expected '-100', got '%d'", args.E)
	}
}

type argsUInts struct {
	A uint   `argo:"short=a"`
	B uint8  `argo:"long=b"`
	C uint16 `argo:"env=C_VALUE"`
	D uint32 `argo:"short,default=3"`
	E uint64 `argo:"positional"`
}

func TestAllUInts(t *testing.T) {
	if err := os.Setenv("C_VALUE", "55555"); err != nil {
		t.Fatalf("failed to set env: %s", err)
	}
	os.Args = []string{"test", "-a", "1", "--b", "250", "1"}

	args := argsUInts{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A != 1 {
		t.Fatalf("expected '1', got '%d'", args.A)
	}
	if args.B != 250 {
		t.Fatalf("expected '250', got '%d'", args.B)
	}
	if args.C != 55555 {
		t.Fatalf("expected '55555', got '%d'", args.C)
	}
	if args.D != 3 {
		t.Fatalf("expected '3', got '%d'", args.D)
	}
	if args.E != 1 {
		t.Fatalf("expected '1', got '%d'", args.E)
	}
}

type argsFloats struct {
	A float32 `argo:"short=a"`
	B float64 `argo:"long=b"`
}

func TestAllFloats(t *testing.T) {
	os.Args = []string{"test", "-a", "1.5", "--b", "-0.05"}
	args := argsFloats{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A != 1.5 {
		t.Fatalf("expected '1.5', got '%f'", args.A)
	}
	if args.B != -0.05 {
		t.Fatalf("expected '-0.05', got '%f'", args.B)
	}
}

type argsBool struct {
	A bool `argo:"short=a"`
}

type argsBoolPositional struct {
	A bool `argo:"positional"`
}

func TestBool(t *testing.T) {
	os.Args = []string{"test", "-a"}
	args0 := argsBool{}
	if err := Parse(&args0); err != nil {
		t.Fatal(err)
	}
	if !args0.A {
		t.Fatal("expected 'A' to be true")
	}

	os.Args = []string{"test", "false"}
	args := argsBoolPositional{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A {
		t.Fatal("expected 'A' to be false")
	}

	os.Args = []string{"test", "1"}
	args = argsBoolPositional{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if !args.A {
		t.Fatal("expected 'A' to be true")
	}

	os.Args = []string{"test", "INVALID"}
	args = argsBoolPositional{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsInterface struct {
	A interface{} `argo:"long=longName"`
}

func TestInterface(t *testing.T) {
	os.Args = []string{"test", "--longName", "123,456"}
	args := argsInterface{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A != "123,456" {
		t.Fatalf("expected '123,456', got '%s'", args.A)
	}
}

type CustomType struct {
	Value  string
	Number int
}

func customTypeSetter(s string, value reflect.Value) error {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return newError(fmt.Sprintf("(%s) (%s)", errCouldNotSet, s))
	}
	value.Field(0).SetString(parts[0])

	num, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return newError(fmt.Sprintf("(%s) (%s)", errCouldNotSet, s))
	}
	value.Field(1).SetInt(num)
	return nil
}

type argsCustom struct {
	A CustomType `argo:"short=a"`
}

func TestCustomSetter(t *testing.T) {
	if err := RegisterSetter(CustomType{}, customTypeSetter); err != nil {
		t.Fatalf("failed to register setter: %s", err)
	}

	// duplicate
	if err := RegisterSetter(CustomType{}, customTypeSetter); err == nil {
		t.Fatal("expected error")
	}

	os.Args = []string{"test", "-a", "name,22"}
	args := argsCustom{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A.Value != "name" {
		t.Fatalf("expected 'name', got '%s'", args.A.Value)
	}
	if args.A.Number != 22 {
		t.Fatalf("expected '22', got '%d'", args.A.Number)
	}

}

type argsPositionalPlacement struct {
	A string `argo:"short=a"`
	B int    `argo:"positional"`
	C string `argo:"long=c"`
}

func TestPositionalPlacement(t *testing.T) {
	os.Args = []string{"test", "-a", "name", "123", "--c", "value"}
	args := argsPositionalPlacement{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	os.Args = []string{"test", "-a", "name", "--c", "value", "123"}
	args = argsPositionalPlacement{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.A != "name" {
		t.Fatalf("expected 'name', got '%s'", args.A)
	}
	if args.B != 123 {
		t.Fatalf("expected '123', got '%d'", args.B)
	}
	if args.C != "value" {
		t.Fatalf("expected 'value', got '%s'", args.C)
	}

	os.Args = []string{"test", "-a", "name", "123"}
	args = argsPositionalPlacement{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}

	os.Args = []string{"test", "-a", "name"}
	args = argsPositionalPlacement{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	os.Args = []string{"test", "-a", "name", "--c", "value", "456", "789", "123"}
	args = argsPositionalPlacement{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsNotExported struct {
	a string `argo:"short=a"`
}

func TestNotExported(t *testing.T) {
	os.Args = []string{"test", "-a", "name"}
	args := argsNotExported{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsInvalidDefault struct {
	A bool    `argo:"short=a,default=345"`
	B float32 `argo:"env=B,default=def"`
	D uint64  `argo:"long=abc,default=abc"`
	C int8    `argo:"positional,default=240"`
}

func TestInvalidDefault(t *testing.T) {
	// B
	os.Args = []string{"test", "-a", "--abc", "2", "1"}
	args := argsInvalidDefault{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	// C
	os.Args = []string{"test", "-a", "--abc", "2"}
	if err := os.Setenv("B", "123"); err != nil {
		t.Fatalf("failed to set env: %s", err)
	}

	args = argsInvalidDefault{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	// A
	os.Args = []string{"test", "--abc", "2", "1"}
	args = argsInvalidDefault{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	// D
	os.Args = []string{"test", "-a", "3"}
	args = argsInvalidDefault{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsDuplicates struct {
	A string `argo:"short=a"`
	B string `argo:"short=a"`
}

type argsDuplicates2 struct {
	A string `argo:"long=abc"`
	B string `argo:"long=abc"`
}

func TestDuplicateName(t *testing.T) {
	os.Args = []string{"test", "-a", "name"}
	args := argsDuplicates{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	os.Args = []string{"test", "--abc", "name"}
	args2 := argsDuplicates2{}
	if err := Parse(&args2); err == nil {
		t.Fatal("expected error")
	}
}

type argsInvalidEnv struct {
	A int `argo:"env=ABC"`
	B int `argo:"short"`
}

func TestInvalidEnv(t *testing.T) {
	if err := os.Setenv("ABC", "invalid"); err != nil {
		t.Fatalf("failed to set env: %s", err)
	}
	os.Args = []string{"test"}

	args := argsInvalidEnv{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}

	if err := os.Setenv("ABC", "123"); err != nil {
		t.Fatalf("failed to set env: %s", err)
	}
	os.Args = []string{"test", "-b", "abc"}

	args = argsInvalidEnv{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsNeitherShortNorLong struct {
	Abc string `argo:"required"`
}

func TestNeitherShortNorLong(t *testing.T) {
	os.Args = []string{"test", "-a", "name"}
	args := argsNeitherShortNorLong{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Abc != "name" {
		t.Fatalf("expected 'name', got '%s'", args.Abc)
	}

	os.Args = []string{"test", "--abc", "name"}
	args = argsNeitherShortNorLong{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if args.Abc != "name" {
		t.Fatalf("expected 'name', got '%s'", args.Abc)
	}
}

type invalidFieldType struct {
	A int
}

type argsInvalidType struct {
	B int              `argo:"short"`
	C invalidFieldType `argo:"short"`
}

func TestUnsupportedType(t *testing.T) {
	os.Args = []string{"test", "-b", "123", "-c", "456"}
	args := argsInvalidType{}
	if err := Parse(&args); err == nil {
		t.Fatal("expected error")
	}
}

type argsPointers struct {
	A   *string `argo:"short=a"`
	B   *int    `argo:"short=b"`
	XYZ *bool   `argo:"long"`
}

func TestPointers(t *testing.T) {
	os.Args = []string{"test", "-a", "name", "-b", "123"}
	args := argsPointers{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
	if *args.A != "name" {
		t.Fatalf("expected 'name', got '%s'", *args.A)
	}
	if *args.B != 123 {
		t.Fatalf("expected '123', got '%d'", *args.B)
	}
	if args.XYZ != nil {
		t.Fatal("expected 'XYZ' to be nil")
	}
}
