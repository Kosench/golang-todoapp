package core_postgres_pool

import "errors"

var (
	ErrNoRows                  = errors.New("no rows")
	ErrViolatesForeignKey      = errors.New("violates foreign key")
	ErrViolatesCheckConstraint = errors.New("violates check constraint")
	ErrUnknown                 = errors.New("unknown")
)
