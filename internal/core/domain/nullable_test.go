package domain_test

import (
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func TestNullable_Semantics(t *testing.T) {
	// 1. "Поле не было передано" — Set=false, Value=nil
	notSent := domain.Nullable[string]{}
	if notSent.Set {
		t.Error("expected Set=false for zero-value Nullable")
	}
	if notSent.Value != nil {
		t.Error("expected Value=nil for zero-value Nullable")
	}

	// 2. "Поле передано со значением" — Set=true, Value=&"hello"
	v := "hello"
	withValue := domain.Nullable[string]{Value: &v, Set: true}
	if !withValue.Set {
		t.Error("expected Set=true")
	}
	if withValue.Value == nil || *withValue.Value != "hello" {
		t.Errorf("expected Value='hello', got %v", withValue.Value)
	}

	// 3. "Поле передано со значением null (удаление)" — Set=true, Value=nil
	withNull := domain.Nullable[string]{Value: nil, Set: true}
	if !withNull.Set {
		t.Error("expected Set=true for explicit null")
	}
	if withNull.Value != nil {
		t.Error("expected Value=nil for explicit null")
	}
}
