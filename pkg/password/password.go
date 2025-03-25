package password

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultMinLength = 8
	defaultMaxLength = 72 // bcrypt max password length
	PasswordCost     = 14 // Reasonable default for most applications
)

type Validator struct {
	minLength uint8
	maxLength uint8
	hasUpper  bool
	hasLower  bool
	hasNumber bool
	hasSymbol bool
}

// Option defines a function type for configuring the validator
type Option func(*Validator)

// WithMinLength sets the minimum password length
func WithMinLength(length uint8) Option {
	return func(v *Validator) {
		v.minLength = length
	}
}

// WithMaxLength sets the maximum password length
func WithMaxLength(length uint8) Option {
	return func(v *Validator) {
		v.maxLength = length
	}
}

// WithUppercase requires uppercase letters
func WithUppercase(required bool) Option {
	return func(v *Validator) {
		v.hasUpper = required
	}
}

// WithLowercase requires lowercase letters
func WithLowercase(required bool) Option {
	return func(v *Validator) {
		v.hasLower = required
	}
}

// WithNumbers requires numeric characters
func WithNumbers(required bool) Option {
	return func(v *Validator) {
		v.hasNumber = required
	}
}

// WithSymbols requires special characters
func WithSymbols(required bool) Option {
	return func(v *Validator) {
		v.hasSymbol = required
	}
}

// NewValidator creates a password validator with sensible defaults
func NewValidator(opts ...Option) *Validator {
	// Default settings for a reasonably secure password
	v := &Validator{
		minLength: defaultMinLength,
		maxLength: defaultMaxLength,
		hasUpper:  true,
		hasLower:  true,
		hasNumber: true,
		hasSymbol: true,
	}

	// Apply any custom options
	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Validate checks if a password meets the requirements
func (v *Validator) Validate(password string) error {
	// Check length
	if err := v.validateLength(password); err != nil {
		return err
	}

	// Check character requirements
	return v.validateCharacters(password)
}

// validateLength checks if the password meets length requirements
func (v *Validator) validateLength(password string) error {
	if len(password) < int(v.minLength) {
		return fmt.Errorf("password must be at least %d characters long", v.minLength)
	}

	if v.maxLength > 0 && len(password) > int(v.maxLength) {
		return fmt.Errorf("password cannot be longer than %d characters", v.maxLength)
	}

	return nil
}

// validateCharacters checks if the password meets character class requirements
func (v *Validator) validateCharacters(password string) error {
	requirements := v.getCharacterRequirements()
	if len(requirements) == 0 {
		return nil
	}

	counts := countCharacterTypes(password)
	var missingReqs []string

	if v.hasUpper && counts.upper == 0 {
		missingReqs = append(missingReqs, "at least one uppercase letter")
	}

	if v.hasLower && counts.lower == 0 {
		missingReqs = append(missingReqs, "at least one lowercase letter")
	}

	if v.hasNumber && counts.number == 0 {
		missingReqs = append(missingReqs, "at least one number")
	}

	if v.hasSymbol && counts.symbol == 0 {
		missingReqs = append(missingReqs, "at least one special character")
	}

	if len(missingReqs) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(missingReqs, ", "))
	}

	return nil
}

// getCharacterRequirements returns a list of enabled character requirements
func (v *Validator) getCharacterRequirements() []string {
	var requirements []string

	if v.hasUpper {
		requirements = append(requirements, "uppercase")
	}

	if v.hasLower {
		requirements = append(requirements, "lowercase")
	}

	if v.hasNumber {
		requirements = append(requirements, "number")
	}

	if v.hasSymbol {
		requirements = append(requirements, "symbol")
	}

	return requirements
}

// charCounts tracks the counts of different character types
type charCounts struct {
	upper  int
	lower  int
	number int
	symbol int
}

// countCharacterTypes counts the occurrence of each character type
func countCharacterTypes(password string) charCounts {
	var counts charCounts

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			counts.upper++
		case unicode.IsLower(char):
			counts.lower++
		case unicode.IsDigit(char):
			counts.number++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			counts.symbol++
		}
	}

	return counts
}

// Hash creates a bcrypt hash from a password
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password against a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
