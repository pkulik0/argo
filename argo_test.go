package argo_test

import (
	"github.com/pkulik0/argo"
	"testing"
)

func TestEmpty(t *testing.T) {
	var args struct{}
	if err := argo.Parse(&args); err != nil {
		t.Fatal(err)
	}
}

func TestNotPointer(t *testing.T) {
	if err := argo.Parse(struct{}{}); err == nil {
		t.Fatal("expected error")
	}
}

type argsSimple struct {
	address string `argo:"short=a;long=addr;help=Address to connect to"`
}

func TestSimple(t *testing.T) {
	var args argsSimple
	if err := argo.Parse(&args); err != nil {
		t.Fatal(err)
	}
}
