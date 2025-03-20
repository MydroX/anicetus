//revive:disable:add-constant

package password

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	// Test cases
	tests := []struct {
		name      string
		password  string
		validator *Validator
		wantErr   bool
	}{
		{
			name:      "Valid password",
			password:  "Test1234!",
			validator: NewValidator(),
			wantErr:   false,
		},
		{
			name:      "Too short",
			password:  "Short1!",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:      "No uppercase",
			password:  "password123!",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:      "No lowercase",
			password:  "PASSWORD123!",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:      "No number",
			password:  "Password!",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:      "No symbol",
			password:  "Password123",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:      "Too long",
			password:  "Password123!Password123!Password123!Password123!Password123!Password123!Password123!Password123!",
			validator: NewValidator(),
			wantErr:   true,
		},
		{
			name:     "Custom validator - valid",
			password: "password", // Only lowercase
			validator: NewValidator(
				WithMinLength(6),
				WithUppercase(false),
				WithNumbers(false),
				WithSymbols(false),
			),
			wantErr: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashAndCheck(t *testing.T) {
	password := "Test1234!"

	// Test hashing
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	// Test correct password check
	if !CheckPasswordHash(password, hash) {
		t.Errorf("CheckPasswordHash() = false, want true for correct password")
	}

	// Test incorrect password check
	if CheckPasswordHash("WrongPassword1!", hash) {
		t.Errorf("CheckPasswordHash() = true, want false for incorrect password")
	}
}

func TestOptionsConfiguration(t *testing.T) {
	// Test custom configurations
	var maxLength uint8 = 50
	var minLength uint8 = 10

	tests := []struct {
		name      string
		options   []Option
		checkFunc func(*Validator) bool
	}{
		{
			name:    "Set minimum length",
			options: []Option{WithMinLength(minLength)},
			checkFunc: func(v *Validator) bool {
				return v.minLength == minLength
			},
		},
		{
			name:    "Set maximum length",
			options: []Option{WithMaxLength(maxLength)},
			checkFunc: func(v *Validator) bool {
				return v.maxLength == maxLength
			},
		},
		{
			name:    "Disable uppercase",
			options: []Option{WithUppercase(false)},
			checkFunc: func(v *Validator) bool {
				return !v.hasUpper
			},
		},
		{
			name:    "Disable lowercase",
			options: []Option{WithLowercase(false)},
			checkFunc: func(v *Validator) bool {
				return !v.hasLower
			},
		},
		{
			name:    "Disable numbers",
			options: []Option{WithNumbers(false)},
			checkFunc: func(v *Validator) bool {
				return !v.hasNumber
			},
		},
		{
			name:    "Disable symbols",
			options: []Option{WithSymbols(false)},
			checkFunc: func(v *Validator) bool {
				return !v.hasSymbol
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.options...)
			if !tt.checkFunc(validator) {
				t.Errorf("%s: option not applied correctly", tt.name)
			}
		})
	}
}
