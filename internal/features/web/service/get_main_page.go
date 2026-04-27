package web_service

import (
	"fmt"
	"path"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

// GetMainPage возвращает содержимое главной HTML-страницы.
// Путь к файлу формируется относительно projectRoot, переданного через конструктор,
// чтобы приложение работало корректно независимо от рабочей директории запуска.
func (s *WebService) GetMainPage() (domain.File, error) {
	htmlFilePath := path.Join(
		s.projectRoot,
		"/public/index.html",
	)

	htmlFile, err := s.webRepository.GetFile(htmlFilePath)
	if err != nil {
		return domain.File{}, fmt.Errorf("get file from repository: %w", err)
	}

	return htmlFile, nil
}
