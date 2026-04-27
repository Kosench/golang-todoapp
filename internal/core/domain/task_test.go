package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
)

func ptr[T any](v T) *T {
	return &v
}

func TestTaskValidate(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	before := now.Add(-time.Hour)

	tests := []struct {
		name    string
		task    domain.Task
		wantErr bool
	}{
		{
			name:    "valid task without description",
			task:    domain.NewTask(1, 1, "Buy groceries", nil, false, now, nil, 1),
			wantErr: false,
		},
		{
			name:    "valid task with description",
			task:    domain.NewTask(1, 1, "Buy groceries", ptr("Milk, bread, eggs"), false, now, nil, 1),
			wantErr: false,
		},
		{
			name: "valid completed task",
			// Завершённая задача: CompletedAt > CreatedAt — всё корректно.
			task:    domain.NewTask(1, 1, "Done task", nil, true, now, &later, 1),
			wantErr: false,
		},
		{
			name:    "valid task with min title length (1 char)",
			task:    domain.NewTask(1, 1, "X", nil, false, now, nil, 1),
			wantErr: false,
		},
		{
			name: "valid task with max title length (100 chars)",
			// strings.Repeat("a", 100) создаёт строку из 100 символов
			task:    domain.NewTask(1, 1, strings.Repeat("a", 100), nil, false, now, nil, 1),
			wantErr: false,
		},
		{
			name: "valid task with unicode title",
			// Проверяем, что валидация считает символы (руны), а не байты.
			// "Привет" — 6 рун, но 12 байт.
			task:    domain.NewTask(1, 1, "Привет", nil, false, now, nil, 1),
			wantErr: false,
		},

		// --- Негативные кейсы: Title ---
		{
			name:    "invalid: empty title",
			task:    domain.NewTask(1, 1, "", nil, false, now, nil, 1),
			wantErr: true,
		},
		{
			name:    "invalid: title too long (101 chars)",
			task:    domain.NewTask(1, 1, strings.Repeat("a", 101), nil, false, now, nil, 1),
			wantErr: true,
		},

		// --- Негативные кейсы: Description ---
		{
			name: "invalid: empty description (non-nil but empty string)",
			// Description = ptr("") → указатель на пустую строку, длина 0 → < 1
			task:    domain.NewTask(1, 1, "Task", ptr(""), false, now, nil, 1),
			wantErr: true,
		},
		{
			name:    "invalid: description too long (1001 chars)",
			task:    domain.NewTask(1, 1, "Task", ptr(strings.Repeat("d", 1001)), false, now, nil, 1),
			wantErr: true,
		},

		// --- Негативные кейсы: Completed / CompletedAt ---
		{
			name: "invalid: completed but CompletedAt is nil",
			// Completed=true, но CompletedAt=nil — нарушение бизнес-правила.
			task:    domain.NewTask(1, 1, "Task", nil, true, now, nil, 1),
			wantErr: true,
		},
		{
			name: "invalid: completed but CompletedAt before CreatedAt",
			// Нельзя завершить задачу раньше, чем она была создана.
			task:    domain.NewTask(1, 1, "Task", nil, true, now, &before, 1),
			wantErr: true,
		},
		{
			name: "invalid: not completed but CompletedAt is set",
			// Completed=false, но CompletedAt != nil — противоречие.
			task:    domain.NewTask(1, 1, "Task", nil, false, now, &later, 1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()

			// Проверяем, совпадает ли факт наличия ошибки с ожиданием
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Дополнительно: если ошибка есть, проверяем, что это ErrInvalidArgument.
			// Это гарантирует, что валидация возвращает правильный тип ошибки,
			// а не, например, случайную ошибку из другого места.
			if err != nil && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Errorf("expected ErrInvalidArgument in chain, got: %v", err)
			}
		})
	}
}

func TestTaskPatchValidate(t *testing.T) {
	tests := []struct {
		name    string
		patch   domain.TaskPatch
		wantErr bool
	}{
		{
			name: "valid: patch title",
			patch: domain.NewTaskPatch(
				// Set=true, Value=&"New title" → меняем title
				domain.Nullable[string]{Value: ptr("New title"), Set: true},
				domain.Nullable[string]{}, // Set=false → не трогаем description
				domain.Nullable[bool]{},   // Set=false → не трогаем completed
			),
			wantErr: false,
		},
		{
			name: "valid: nothing set",
			// Пустой патч — ничего не меняем. Это валидно.
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: false,
		},
		{
			name: "valid: set description to null (delete it)",
			// Set=true, Value=nil → удаляем описание. Это допустимо для Description.
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: nil, Set: true},
				domain.Nullable[bool]{},
			),
			wantErr: false,
		},
		{
			name: "invalid: title set to null",
			// Title — обязательное поле, нельзя обнулить.
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Value: nil, Set: true},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: true,
		},
		{
			name: "invalid: completed set to null",
			// Completed — обязательное поле, нельзя обнулить.
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{},
				domain.Nullable[bool]{Value: nil, Set: true},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TaskPatch.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskApplyPatch(t *testing.T) {
	now := time.Now()
	patchTime := now.Add(2 * time.Hour) // время, которое передаём в ApplyPatch как "now"

	// Базовая задача, от которой будем отталкиваться в каждом тесте.
	// Создаём функцию-фабрику, чтобы каждый тест получал свежую копию.
	baseTask := func() domain.Task {
		return domain.NewTask(
			1, 1,
			"Original title",
			ptr("Original description"),
			false, now, nil, 1,
		)
	}

	tests := []struct {
		name      string
		patch     domain.TaskPatch
		wantErr   bool
		wantTitle string // ожидаемый title после патча
		wantCompl bool   // ожидаемый Completed после патча
	}{
		{
			name: "patch title only",
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Value: ptr("Updated title"), Set: true},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr:   false,
			wantTitle: "Updated title",
			wantCompl: false, // не менялось
		},
		{
			name: "patch completed to true",
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{},
				domain.Nullable[bool]{Value: ptr(true), Set: true},
			),
			wantErr:   false,
			wantTitle: "Original title", // не менялось
			wantCompl: true,
		},
		{
			name: "delete description (set to null)",
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: nil, Set: true}, // description → null
				domain.Nullable[bool]{},
			),
			wantErr:   false,
			wantTitle: "Original title",
			wantCompl: false,
		},
		{
			name: "invalid patch: title to empty string",
			// Title = "" → после ApplyPatch, tmp.Validate() вернёт ошибку,
			// потому что len([]rune("")) = 0 < 1.
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Value: ptr(""), Set: true},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: true,
		},
		{
			name: "invalid patch: title set to null",
			// Ошибка на этапе patch.Validate(), до tmp.Validate().
			patch: domain.NewTaskPatch(
				domain.Nullable[string]{Value: nil, Set: true},
				domain.Nullable[string]{},
				domain.Nullable[bool]{},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := baseTask() // свежая копия для каждого теста

			err := task.ApplyPatch(tt.patch, patchTime)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ApplyPatch() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Если ошибки нет — проверяем, что поля изменились правильно
			if !tt.wantErr {
				if task.Title != tt.wantTitle {
					t.Errorf("Title = %q, want %q", task.Title, tt.wantTitle)
				}
				if task.Completed != tt.wantCompl {
					t.Errorf("Completed = %v, want %v", task.Completed, tt.wantCompl)
				}
				// Если пометили как завершённую, CompletedAt должен быть равен patchTime
				if tt.wantCompl && (task.CompletedAt == nil || !task.CompletedAt.Equal(patchTime)) {
					t.Errorf("CompletedAt = %v, want %v", task.CompletedAt, patchTime)
				}
			}
		})
	}
}

// Отдельный тест: ApplyPatch не должен мутировать оригинал при ошибке.
func TestTaskApplyPatch_DoesNotMutateOnError(t *testing.T) {
	now := time.Now()
	task := domain.NewTask(1, 1, "Original", nil, false, now, nil, 1)

	// Пытаемся применить невалидный патч (title → null)
	patch := domain.NewTaskPatch(
		domain.Nullable[string]{Value: nil, Set: true},
		domain.Nullable[string]{},
		domain.Nullable[bool]{},
	)

	err := task.ApplyPatch(patch, now)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Оригинал не должен измениться
	if task.Title != "Original" {
		t.Errorf("task.Title was mutated to %q, expected 'Original'", task.Title)
	}
}

func TestTaskCompletionDuration(t *testing.T) {
	now := time.Now()
	oneHourLater := now.Add(time.Hour)

	tests := []struct {
		name         string
		task         domain.Task
		wantNil      bool
		wantDuration time.Duration
	}{
		{
			name:    "not completed — returns nil",
			task:    domain.NewTask(1, 1, "Task", nil, false, now, nil, 1),
			wantNil: true,
		},
		{
			name: "completed but CompletedAt is nil — returns nil",
			// Такое состояние невалидно (Validate вернёт ошибку),
			// но CompletionDuration() должен быть безопасен и не паниковать.
			task:    domain.NewTask(1, 1, "Task", nil, true, now, nil, 1),
			wantNil: true,
		},
		{
			name:         "completed with CompletedAt — returns duration",
			task:         domain.NewTask(1, 1, "Task", nil, true, now, &oneHourLater, 1),
			wantNil:      false,
			wantDuration: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.CompletionDuration()

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", *got)
				}
				return // дальше проверять нечего
			}

			if got == nil {
				t.Fatal("expected non-nil duration, got nil")
			}
			if *got != tt.wantDuration {
				t.Errorf("duration = %v, want %v", *got, tt.wantDuration)
			}
		})
	}
}
