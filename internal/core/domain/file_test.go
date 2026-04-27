package domain_test

import (
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func TestNewFile(t *testing.T) {
	content := []byte("<html>Hello</html>")
	file := domain.NewFile(content)

	got := file.Buffer()
	if len(got) != len(content) {
		t.Fatalf("Buffer() len = %d, want %d", len(got), len(content))
	}

	for i := range content {
		if got[i] != content[i] {
			t.Fatalf("Buffer()[%d] = %d, want %d", i, got[i], content[i])
		}
	}
}

func TestNewFile_Empty(t *testing.T) {
	file := domain.NewFile([]byte{})

	if len(file.Buffer()) != 0 {
		t.Errorf("expected empty buffer, got len=%d", len(file.Buffer()))
	}

}
