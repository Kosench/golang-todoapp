package domain

import (
	"fmt"
	"time"

	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
)

type Task struct {
	ID      int
	Version int

	Title        string
	Description  *string
	Completed    bool
	CreatedAt    time.Time
	CompletedAt  *time.Time
	AuthorUserID int
}

func NewTask(
	id int,
	version int,
	title string,
	description *string,
	completed bool,
	createdAt time.Time,
	completedAt *time.Time,
	authorUserID int,
) Task {
	return Task{
		ID:           id,
		Version:      version,
		Title:        title,
		Description:  description,
		Completed:    completed,
		CreatedAt:    createdAt,
		CompletedAt:  completedAt,
		AuthorUserID: authorUserID,
	}
}

func NewTaskUninitialized(title string, description *string, authorUserID int) Task {
	return NewTask(
		UninitializedID,
		UninitializedVersion,
		title,
		description,
		false,
		time.Now(),
		nil,
		authorUserID,
	)
}

func (t *Task) Validate() error {
	titleLen := len([]rune(t.Title))
	if titleLen < 1 || titleLen > 100 {
		return fmt.Errorf(
			"invalid title len: %d: %w",
			titleLen,
			core_errors.ErrInvalidArgument)
	}

	if t.Description != nil {
		descriptionLen := len([]rune(*t.Description))
		if descriptionLen < 1 || descriptionLen > 100 {
			return fmt.Errorf(
				"description title len: %d: %w",
				titleLen,
				core_errors.ErrInvalidArgument)
		}
	}

	if t.Completed {
		if t.CompletedAt == nil {
			return fmt.Errorf("`CompletedAt` cant be nil if `Completed`==`true`: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		if t.CompletedAt.Before(t.CreatedAt) {
			return fmt.Errorf("`CompletedAt` cant be before `CreatedAt`: %w",
				core_errors.ErrInvalidArgument,
			)
		}
	} else {
		if t.CompletedAt != nil {
			return fmt.Errorf("`CompletedAt` must be `nil` if Completed==false: %w",
				core_errors.ErrInvalidArgument,
			)
		}
	}

	return nil
}
