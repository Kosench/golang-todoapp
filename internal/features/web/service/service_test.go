package web_service_test

import (
	"errors"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	web_service "github.com/Kosench/golang-todoapp/internal/features/web/service"
)

type MockWebRepository struct {
	GetFileFunc func(filePath string) (domain.File, error)
}

func (m *MockWebRepository) GetFile(filePath string) (domain.File, error) {
	return m.GetFileFunc(filePath)
}

func TestWebService_GetMainPage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		htmlContent := []byte("<html><body>Hello</body></html>")

		repo := &MockWebRepository{
			GetFileFunc: func(filePath string) (domain.File, error) {
				// Проверяем, что путь сформирован правильно
				expectedPath := "/app/public/index.html"
				if filePath != expectedPath {
					t.Errorf("GetFile called with path=%q, want %q", filePath, expectedPath)
				}
				return domain.NewFile(htmlContent), nil
			},
		}

		// projectRoot = "/app" → ожидаем путь "/app/public/index.html"
		svc := web_service.NewWebService(repo, "/app")
		got, err := svc.GetMainPage()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Buffer()) != len(htmlContent) {
			t.Errorf("Buffer len = %d, want %d", len(got.Buffer()), len(htmlContent))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		repo := &MockWebRepository{
			GetFileFunc: func(filePath string) (domain.File, error) {
				return domain.File{}, core_errors.ErrNotFound
			},
		}

		svc := web_service.NewWebService(repo, "/app")
		_, err := svc.GetMainPage()

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockWebRepository{
			GetFileFunc: func(filePath string) (domain.File, error) {
				return domain.File{}, errors.New("permission denied")
			},
		}

		svc := web_service.NewWebService(repo, "/app")
		_, err := svc.GetMainPage()

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
