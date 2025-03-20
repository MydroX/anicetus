// nolint:all
package argon2id

import (
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	t.Run("Generate hash with default params", func(t *testing.T) {
		params := New()
		password := "test_password"

		hash, err := Hash(password, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Validate hash format
		parts := strings.Split(hash, "$")
		if len(parts) != encodedHashParts {
			t.Errorf("Hash format invalid, got %d parts, want %d", len(parts), encodedHashParts)
		}

		if parts[1] != "argon2id" {
			t.Errorf("Hash algorithm = %s, want argon2id", parts[1])
		}

		if parts[2] != "v=19" {
			t.Errorf("Version = %s, want v=19", parts[2])
		}

		// Check params part contains memory, iterations, parallelism
		paramPart := parts[3]
		if !strings.Contains(paramPart, "m=") ||
			!strings.Contains(paramPart, "t=") ||
			!strings.Contains(paramPart, "p=") {
			t.Errorf("Missing parameters in hash: %s", paramPart)
		}

		// Check salt and hash parts exist and aren't empty
		if parts[4] == "" {
			t.Errorf("Salt part is empty")
		}
		if parts[5] == "" {
			t.Errorf("Hash part is empty")
		}
	})

	t.Run("Generate hash with custom params", func(t *testing.T) {
		params := New(
			Memory(128*1024),
			Iterations(4),
			Parallelism(4),
			KeyLength(32),
			SaltLength(16),
		)
		password := "test_password"

		hash, err := Hash(password, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Check the parameters were included in hash
		if !strings.Contains(hash, "m=131072") {
			t.Errorf("Hash doesn't contain correct memory param: %s", hash)
		}
		if !strings.Contains(hash, "t=4") {
			t.Errorf("Hash doesn't contain correct iterations param: %s", hash)
		}
		if !strings.Contains(hash, "p=4") {
			t.Errorf("Hash doesn't contain correct parallelism param: %s", hash)
		}
	})

	t.Run("Different passwords produce different hashes", func(t *testing.T) {
		params := New()
		hash1, _ := Hash("password1", params)
		hash2, _ := Hash("password2", params)

		if hash1 == hash2 {
			t.Errorf("Different passwords produced the same hash")
		}
	})

	t.Run("Same password produces different hashes (salt randomness)", func(t *testing.T) {
		params := New()
		password := "same_password"

		hash1, _ := Hash(password, params)
		hash2, _ := Hash(password, params)

		if hash1 == hash2 {
			t.Errorf("Same password produced the same hash, which indicates salt isn't random")
		}
	})
}

func TestVerify(t *testing.T) {
	t.Run("Correct password should verify", func(t *testing.T) {
		password := "correct_password"
		params := New()

		// First generate a hash
		hash, err := Hash(password, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Then verify it
		valid, err := Verify(password, hash)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}

		if !valid {
			t.Errorf("Verify() = %v, want true for correct password", valid)
		}
	})

	t.Run("Incorrect password should not verify", func(t *testing.T) {
		correctPassword := "correct_password"
		wrongPassword := "wrong_password"
		params := New()

		// First generate a hash with the correct password
		hash, err := Hash(correctPassword, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Then verify with the wrong password
		valid, err := Verify(wrongPassword, hash)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}

		if valid {
			t.Errorf("Verify() = %v, want false for incorrect password", valid)
		}
	})

	t.Run("Invalid hash format should return error", func(t *testing.T) {
		_, err := Verify("password", "invalid_hash_format")

		if err == nil {
			t.Errorf("Expected error for invalid hash format, got nil")
		}
	})

	t.Run("Hash with incorrect parameter format should return error", func(t *testing.T) {
		// Create a malformed hash with invalid parameter format
		malformedHash := "$argon2id$v=19$invalid_params$c2FsdA$aGFzaA"

		_, err := Verify("password", malformedHash)

		if err == nil {
			t.Errorf("Expected error for invalid parameter format, got nil")
		}
	})

	t.Run("Hash with invalid base64 salt should return error", func(t *testing.T) {
		// Create a malformed hash with invalid base64 salt
		malformedHash := "$argon2id$v=19$m=65536,t=3,p=2$invalid!!!$aGFzaA"

		_, err := Verify("password", malformedHash)

		if err == nil {
			t.Errorf("Expected error for invalid base64 salt, got nil")
		}
	})

	t.Run("Hash with invalid base64 hash should return error", func(t *testing.T) {
		// Create a malformed hash with invalid base64 hash
		malformedHash := "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$invalid!!!"

		_, err := Verify("password", malformedHash)

		if err == nil {
			t.Errorf("Expected error for invalid base64 hash, got nil")
		}
	})

	t.Run("Verify works with custom parameters", func(t *testing.T) {
		password := "custom_params_password"
		params := New(
			Memory(128*1024),
			Iterations(4),
			Parallelism(4),
		)

		// Generate hash with custom parameters
		hash, err := Hash(password, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Verify it
		valid, err := Verify(password, hash)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}

		if !valid {
			t.Errorf("Verify() = %v, want true for password with custom params", valid)
		}
	})

	t.Run("Verify with very low memory setting", func(t *testing.T) {
		password := "low_memory_password"
		params := New(
			Memory(8), // Very low setting for speed
			Iterations(1),
		)

		// Generate hash with minimum parameters
		hash, err := Hash(password, params)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		// Verify it
		valid, err := Verify(password, hash)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}

		if !valid {
			t.Errorf("Verify() = %v, want true for password with low memory params", valid)
		}
	})
}
