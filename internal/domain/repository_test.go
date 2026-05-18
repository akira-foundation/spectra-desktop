package domain

import (
	"errors"
	"testing"
)

func TestErrProjectNotFound_IsItself(t *testing.T) {
	if !errors.Is(ErrProjectNotFound, ErrProjectNotFound) {
		t.Fatalf("sentinel must match itself")
	}
	if ErrProjectNotFound.Error() != "project not found" {
		t.Fatalf("message drifted: %q", ErrProjectNotFound.Error())
	}
}

func TestErrProjectExists_IsItself(t *testing.T) {
	if !errors.Is(ErrProjectExists, ErrProjectExists) {
		t.Fatalf("sentinel must match itself")
	}
	if ErrProjectExists.Error() != "project already exists" {
		t.Fatalf("message drifted: %q", ErrProjectExists.Error())
	}
}

func TestErr_Distinct(t *testing.T) {
	if errors.Is(ErrProjectNotFound, ErrProjectExists) {
		t.Fatalf("sentinels must be distinct")
	}
}
