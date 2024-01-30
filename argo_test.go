package argo

import (
	"os"
	"reflect"
	"testing"
)

func TestParseAttributesFlags(t *testing.T) {
	attribs := &field{}
	if err := parseAttribute("name", "short=s", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "s" {
		t.Fatalf("expected 's', got '%s'", attribs.short)
	}

	attribs = &field{}
	if err := parseAttribute("name", "short", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "n" {
		t.Fatalf("expected 'n', got '%s'", attribs.short)
	}

	attribs = &field{}
	if err := parseAttribute("name", "long", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "name" {
		t.Fatalf("expected 'name', got '%s'", attribs.long)
	}

	attribs = &field{}
	if err := parseAttribute("name", "long=longname", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "longname" {
		t.Fatalf("expected 'longname', got '%s'", attribs.long)
	}

	attribs = &field{}
	if err := parseAttribute("name", "positional", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.isPositional {
		t.Fatal("expected 'isPositional' to be true")
	}
}

func TestParseAttributesBool(t *testing.T) {
	attribs := &field{}
	if err := parseAttribute("name", "required", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.isRequired {
		t.Fatal("expected 'isRequired' to be true")
	}

	attribs = &field{}
	if err := parseAttribute("name", "required=false", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.isRequired {
		t.Fatal("expected 'isRequired' to be false")
	}
}

func TestParseAttributesString(t *testing.T) {
	attribs := &field{}
	if err := parseAttribute("name", "help=help text", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.help != "help text" {
		t.Fatalf("expected 'help text', got '%s'", attribs.help)
	}

	attribs = &field{}
	if err := parseAttribute("name", "default=123,abc", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.defaultValue != "123,abc" {
		t.Fatalf("expected '123,abc', got '%s'", attribs.defaultValue)
	}

	attribs = &field{}
	if err := parseAttribute("name", "env", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.env != "NAME" {
		t.Fatalf("expected 'env' to be 'NAME', got '%s'", attribs.env)
	}

	attribs = &field{}
	if err := parseAttribute("abc_123", "env=abc_123", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.env != "abc_123" {
		t.Fatalf("expected 'env' to be 'abc_123', got '%s'", attribs.env)
	}
}

func TestParseAttributesInvalid(t *testing.T) {
	attribs := &field{}
	if err := parseAttribute("name", "help", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "short=3", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "short=ab_uu", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "long= eoeeo", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "required=something", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "env=321", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "default", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "default=", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "unknown", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &field{}
	if err := parseAttribute("name", "positional=123", attribs); err == nil {
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

	attribs, err := parseField(fieldValue, structField)
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

	attribs, err = parseField(fieldValue, structField)
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

func TestEmpty(t *testing.T) {
	os.Args = []string{"test"}
	var args struct{}
	if err := Parse(&args); err != nil {
		t.Fatal(err)
	}
}

func TestNotPointer(t *testing.T) {
	if err := Parse(struct{}{}); err == nil {
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
