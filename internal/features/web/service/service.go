// Package web_service содержит сервис для отдачи статических веб-страниц.
package web_service

import "github.com/Kosench/golang-todoapp/internal/core/domain"

// WebService — сервис для работы с веб-страницами приложения.
type WebService struct {
	webRepository WebRepository
	projectRoot   string
}

// WebRepository — интерфейс репозитория для чтения файлов.
type WebRepository interface {
	GetFile(filePath string) (domain.File, error)
}

// NewWebService создаёт сервис с внедрённым репозиторием файловой системы.
// projectRoot — корневая директория проекта для построения путей к статическим файлам.
func NewWebService(
	webRepository WebRepository,
	projectRoot string,
) *WebService {
	return &WebService{
		webRepository: webRepository,
		projectRoot:   projectRoot,
	}
}
