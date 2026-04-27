package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
)

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
	}{
		// --- Позитивные ---
		{
			name:    "valid user without phone",
			user:    domain.NewUser(1, 1, "John Doe", nil),
			wantErr: false,
		},
		{
			name:    "valid user with phone",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+1234567890")),
			wantErr: false,
		},
		{
			name:    "valid: min FullName length (3 chars)",
			user:    domain.NewUser(1, 1, "Bob", nil),
			wantErr: false,
		},
		{
			name:    "valid: max FullName length (100 chars)",
			user:    domain.NewUser(1, 1, strings.Repeat("a", 100), nil),
			wantErr: false,
		},
		{
			name:    "valid: min phone length (10 chars)",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+123456789")),
			wantErr: false,
		},
		{
			name:    "valid: max phone length (15 chars)",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+12345678901234")),
			wantErr: false,
		},
		{
			name:    "valid: unicode FullName",
			user:    domain.NewUser(1, 1, "Иван Иванов", nil),
			wantErr: false,
		},

		// --- Негативные: FullName ---
		{
			name:    "invalid: FullName too short (2 chars)",
			user:    domain.NewUser(1, 1, "AB", nil),
			wantErr: true,
		},
		{
			name:    "invalid: FullName empty",
			user:    domain.NewUser(1, 1, "", nil),
			wantErr: true,
		},
		{
			name:    "invalid: FullName too long (101 chars)",
			user:    domain.NewUser(1, 1, strings.Repeat("a", 101), nil),
			wantErr: true,
		},

		// --- Негативные: PhoneNumber ---
		{
			name:    "invalid: phone too short (9 chars)",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+12345678")),
			wantErr: true,
		},
		{
			name:    "invalid: phone too long (16 chars)",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+123456789012345")),
			wantErr: true,
		},
		{
			name: "invalid: phone without + prefix",
			// Формат требует начала с +, затем только цифры
			user:    domain.NewUser(1, 1, "John Doe", ptr("1234567890")),
			wantErr: true,
		},
		{
			name:    "invalid: phone with letters",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+123456abc")),
			wantErr: true,
		},
		{
			name:    "invalid: phone with spaces",
			user:    domain.NewUser(1, 1, "John Doe", ptr("+123 456 789")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Errorf("expected ErrInvalidArgument in chain, got: %v", err)
			}
		})
	}
}

func TestUserPatchValidate(t *testing.T) {
	tests := []struct {
		name    string
		patch   domain.UserPatch
		wantErr bool
	}{
		{
			name: "valid: patch FullName",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{Value: ptr("New Name"), Set: true},
				domain.Nullable[string]{},
			),
			wantErr: false,
		},
		{
			name: "valid: delete phone (set to null)",
			// PhoneNumber можно удалить — это допустимо бизнес-правилами.
			patch: domain.NewUserPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: nil, Set: true},
			),
			wantErr: false,
		},
		{
			name: "valid: empty patch (nothing set)",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{},
			),
			wantErr: false,
		},
		{
			name: "invalid: FullName set to null",
			// FullName — обязательное поле, нельзя удалить.
			patch: domain.NewUserPatch(
				domain.Nullable[string]{Value: nil, Set: true},
				domain.Nullable[string]{},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserPatch.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserApplyPatch(t *testing.T) {
	baseUser := func() domain.User {
		return domain.NewUser(1, 1, "John Doe", ptr("+1234567890"))
	}

	tests := []struct {
		name         string
		patch        domain.UserPatch
		wantErr      bool
		wantFullName string
		wantPhoneNil bool // ожидаем, что PhoneNumber стал nil
	}{
		{
			name: "patch FullName",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{Value: ptr("Jane Doe"), Set: true},
				domain.Nullable[string]{},
			),
			wantErr:      false,
			wantFullName: "Jane Doe",
			wantPhoneNil: false,
		},
		{
			name: "delete phone number",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: nil, Set: true},
			),
			wantErr:      false,
			wantFullName: "John Doe", // не менялось
			wantPhoneNil: true,
		},
		{
			name: "update phone number",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: ptr("+9876543210"), Set: true},
			),
			wantErr:      false,
			wantFullName: "John Doe",
			wantPhoneNil: false,
		},
		{
			name: "invalid: FullName too short after patch",
			// Меняем FullName на "AB" (2 символа) — Validate() после патча упадёт.
			patch: domain.NewUserPatch(
				domain.Nullable[string]{Value: ptr("AB"), Set: true},
				domain.Nullable[string]{},
			),
			wantErr: true,
		},
		{
			name: "invalid: phone with wrong format after patch",
			patch: domain.NewUserPatch(
				domain.Nullable[string]{},
				domain.Nullable[string]{Value: ptr("not-a-phone"), Set: true},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := baseUser()

			err := user.ApplyPatch(tt.patch)

			if (err != nil) != tt.wantErr {
				t.Fatalf("User.ApplyPatch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if user.FullName != tt.wantFullName {
					t.Errorf("FullName = %q, want %q", user.FullName, tt.wantFullName)
				}
				if tt.wantPhoneNil && user.PhoneNumber != nil {
					t.Errorf("PhoneNumber = %v, want nil", *user.PhoneNumber)
				}
				if !tt.wantPhoneNil && user.PhoneNumber == nil {
					t.Error("PhoneNumber is nil, want non-nil")
				}
			}
		})
	}
}

func TestUserApplyPatch_DoesNotMutateOnError(t *testing.T) {
	user := domain.NewUser(1, 1, "Original Name", ptr("+1234567890"))

	patch := domain.NewUserPatch(
		domain.Nullable[string]{Value: ptr("AB"), Set: true}, // слишком короткое имя
		domain.Nullable[string]{},
	)

	err := user.ApplyPatch(patch)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if user.FullName != "Original Name" {
		t.Errorf("user.FullName was mutated to %q", user.FullName)
	}
}
