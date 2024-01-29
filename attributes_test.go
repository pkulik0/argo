package argo

import "testing"

func TestParseAttributesFlags(t *testing.T) {
	attribs := &fieldAttributes{}
	if err := parseAttribute("name", "short=s", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "s" {
		t.Fatalf("expected 's', got '%s'", attribs.short)
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "short", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.short != "n" {
		t.Fatalf("expected 'n', got '%s'", attribs.short)
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "long", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "name" {
		t.Fatalf("expected 'name', got '%s'", attribs.long)
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "long=longname", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.long != "longname" {
		t.Fatalf("expected 'longname', got '%s'", attribs.long)
	}
}

func TestParseAttributesBool(t *testing.T) {
	attribs := &fieldAttributes{}
	if err := parseAttribute("name", "required", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.isRequired {
		t.Fatal("expected 'isRequired' to be true")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "required=false", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.isRequired {
		t.Fatal("expected 'isRequired' to be false")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "env", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.fromEnv {
		t.Fatal("expected 'fromEnv' to be true")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "env=true", attribs); err != nil {
		t.Fatal(err)
	}
	if !attribs.fromEnv {
		t.Fatal("expected 'fromEnv' to be true")
	}
}

func TestParseAttributesString(t *testing.T) {
	attribs := &fieldAttributes{}
	if err := parseAttribute("name", "help=help text", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.help != "help text" {
		t.Fatalf("expected 'help text', got '%s'", attribs.help)
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "default=123,abc", attribs); err != nil {
		t.Fatal(err)
	}
	if attribs.defaultValue != "123,abc" {
		t.Fatalf("expected '123,abc', got '%s'", attribs.defaultValue)
	}
}

func TestParseAttributesInvalid(t *testing.T) {
	attribs := &fieldAttributes{}
	if err := parseAttribute("name", "help", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "short=3", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "short=ab_uu", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "long= eoeeo", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "required=something", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "env=321", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "default", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "default=", attribs); err == nil {
		t.Fatal("expected error")
	}

	attribs = &fieldAttributes{}
	if err := parseAttribute("name", "unknown", attribs); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseField(t *testing.T) {
	attribs, err := parseField("name", "short=s,long=longname,help=help text,required,env,default=123abc")
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
	if !attribs.fromEnv {
		t.Fatal("expected 'fromEnv' to be true")
	}
	if attribs.defaultValue != "123abc" {
		t.Fatalf("expected '123abc', got '%s'", attribs.defaultValue)
	}

	attribs, err = parseField("name", "short,long=abc123,required=False,default=default value")
	if err != nil {
		t.Fatal(err)
	}
	if attribs.short != "n" {
		t.Fatalf("expected 'n', got '%s'", attribs.short)
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
	if attribs.fromEnv {
		t.Fatal("expected 'fromEnv' to be false")
	}
	if attribs.defaultValue != "default value" {
		t.Fatalf("expected 'default value', got '%s'", attribs.defaultValue)
	}
}
