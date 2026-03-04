package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{ // 0
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901", "22345678901"},
			},
			expectedErr: nil,
		},
		{ // 1
			in: User{
				ID:     "short-id",
				Name:   "John Doe",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{Field: "ID", Err: ErrInvalidLength},
			},
		},
		{ // 2
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    51,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{Field: "Age", Err: ErrMaxExceeded},
			},
		},
		{ // 3
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "invalid-email",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{Field: "Email", Err: ErrNoMatchRegexp},
			},
		},
		{ // 4
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "John Doe",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345"},
			},
			expectedErr: ValidationErrors{
				{Field: "Phones[0]", Err: ErrInvalidLength},
			},
		},
		{ // 5
			in: User{
				ID:     "short",
				Name:   "John Doe",
				Age:    17,
				Email:  "invalid",
				Role:   "superuser",
				Phones: []string{"12345", "67890"},
			},
			expectedErr: ValidationErrors{
				{Field: "ID", Err: ErrInvalidLength},
				{Field: "Age", Err: ErrMinNotReached},
				{Field: "Email", Err: ErrNoMatchRegexp},
				{Field: "Role", Err: ErrNotInSet},
				{Field: "Phones[0]", Err: ErrInvalidLength},
				{Field: "Phones[1]", Err: ErrInvalidLength},
			},
		},
		{ // 6
			in: App{
				Version: "1.2.3",
			},
			expectedErr: nil,
		},
		{ // 7
			in: App{
				Version: "1.2",
			},
			expectedErr: ValidationErrors{
				{Field: "Version", Err: ErrInvalidLength},
			},
		},
		{ // 8
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{ // 9
			in:          "not a struct",
			expectedErr: ErrInvalidStruct,
		},
		{ // 10
			in: struct {
				Field string `validate:"len"`
			}{Field: "test"},
			expectedErr: ErrInvalidRuleFormat,
		},
		{ // 11
			in: struct {
				Field int `validate:"unknown:10"`
			}{Field: 5},
			expectedErr: ErrUnknownRule,
		},
		{ // 12
			in: struct {
				Field int `validate:"max:notanumber"`
			}{Field: 5},
			expectedErr: ErrInvalidParam,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			check(t, tt)
		})
	}
}

func check(t *testing.T, tt struct {
	in          interface{}
	expectedErr error
},
) {
	t.Helper()

	err := Validate(tt.in)

	if tt.expectedErr == nil {
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		return
	}

	if err == nil {
		t.Errorf("expected error %v, got nil", tt.expectedErr)
		return
	}

	if errors.Is(tt.expectedErr, ErrInvalidStruct) ||
		errors.Is(tt.expectedErr, ErrInvalidRuleFormat) ||
		errors.Is(tt.expectedErr, ErrUnknownRule) ||
		errors.Is(tt.expectedErr, ErrInvalidParam) {
		if !errors.Is(err, tt.expectedErr) {
			t.Errorf("expected error %v, got %v", tt.expectedErr, err)
		}
		return
	}

	checkValidationError(t, err, tt.expectedErr)
}

func checkValidationError(t *testing.T, actual, expected error) {
	t.Helper()

	var expectedErrs ValidationErrors
	var actualErrs ValidationErrors

	if !errors.As(expected, &expectedErrs) {
		t.Fatalf("expected error must be ValidationErrors")
	}

	if !errors.As(actual, &actualErrs) {
		t.Fatalf("expected ValidationErrors, got %T", actual)
	}

	if len(actualErrs) != len(expectedErrs) {
		t.Errorf("expected %d errors, got %d", len(expectedErrs), len(actualErrs))
		t.Logf("Expected: %v", expectedErrs)
		t.Logf("Actual: %v", actualErrs)
		return
	}

	for _, expectedVe := range expectedErrs {
		found := false
		for _, actualVe := range actualErrs {
			if actualVe.Field == expectedVe.Field {
				found = true

				if !errors.Is(actualVe.Err, expectedVe.Err) {
					t.Errorf("expected error %v, got %v", expectedVe.Err, actualVe.Err)
				}
				break
			}
		}
		if !found {
			t.Errorf("field %q not found in actual errors", expectedVe.Field)
		}
	}
}
