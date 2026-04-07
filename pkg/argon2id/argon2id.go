package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	DefaultMemory      = 64 * 1024
	DefaultIterations  = 3
	DefaultParallelism = 2
	DefaultSaltLength  = 32
	DefaultKeyLength   = 32

	encodedHashParts = 6

	hashParamPartsCount = 2

	// Security limits
	MaxMemory      = 1024 * 1024 * 1024 // 1GB max
	MaxIterations  = 100                // Reasonable max iterations
	MaxParallelism = 255                // uint8 max
)

type HashParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
	Salt        []byte
	Hash        []byte
}

type ParsedHash struct {
	Params *HashParams
	Salt   []byte
	Hash   []byte
}

// New creates a new HashParams instance with the provided secret key and options
func New(options ...func(*HashParams)) *HashParams {
	h := &HashParams{}

	for _, option := range options {
		option(h)
	}

	if h.Memory == 0 {
		h.Memory = DefaultMemory
	}

	if h.Iterations == 0 {
		h.Iterations = DefaultIterations
	}

	if h.Parallelism == 0 {
		h.Parallelism = DefaultParallelism
	}

	if h.SaltLength == 0 {
		h.SaltLength = DefaultSaltLength
	}

	if h.KeyLength == 0 {
		h.KeyLength = DefaultKeyLength
	}

	return h
}

func Memory(memory uint32) func(*HashParams) {
	return func(h *HashParams) {
		h.Memory = memory
	}
}

func Iterations(iterations uint32) func(*HashParams) {
	return func(h *HashParams) {
		h.Iterations = iterations
	}
}

func Parallelism(parallelism uint8) func(*HashParams) {
	return func(h *HashParams) {
		h.Parallelism = parallelism
	}
}

func SaltLength(saltLength uint32) func(*HashParams) {
	return func(h *HashParams) {
		h.SaltLength = saltLength
	}
}

func KeyLength(keyLength uint32) func(*HashParams) {
	return func(h *HashParams) {
		h.KeyLength = keyLength
	}
}

// Hash generates an encoded hash from a string
func Hash(str string, params *HashParams) (string, error) {
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(str), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify compares a token with an encoded hash
func Verify(token, encodedHash string) (bool, error) {
	// Parse the encoded hash
	parsedHash, err := parseEncodedHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Compute the hash of the input token
	computedHash := argon2.IDKey(
		[]byte(token),
		parsedHash.Salt,
		parsedHash.Params.Iterations,
		parsedHash.Params.Memory,
		parsedHash.Params.Parallelism,
		parsedHash.Params.KeyLength,
	)

	// Compare the computed hash with the stored hash
	return subtle.ConstantTimeCompare(parsedHash.Hash, computedHash) == 1, nil
}

// parseEncodedHash extracts parameters, salt and hash from the encoded hash string
func parseEncodedHash(encodedHash string) (*ParsedHash, error) {
	// Split the encoded hash into parts
	vals := strings.Split(encodedHash, "$")
	if len(vals) != encodedHashParts {
		return nil, errors.New("invalid hash format")
	}

	// Parse parameters
	params, err := parseParams(vals[3])
	if err != nil {
		return nil, err
	}

	// Extract salt
	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, fmt.Errorf("invalid salt: %w", err)
	}

	// Extract hash
	hash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}

	saltLen := len(salt)
	if saltLen > 0 && saltLen <= 1<<32-1 {
		params.SaltLength = uint32(saltLen)
	}

	hashLen := len(hash)
	if hashLen > 0 && hashLen <= 1<<32-1 {
		params.KeyLength = uint32(hashLen)
	}

	return &ParsedHash{
		Params: params,
		Salt:   salt,
		Hash:   hash,
	}, nil
}

// parseParams extracts algorithm parameters from the parameters string
func parseParams(paramsStr string) (*HashParams, error) {
	params := &HashParams{}
	paramParts := strings.Split(paramsStr, ",")

	paramHandlers := map[string]func(string, *HashParams) error{
		"m": parseMemory,
		"t": parseIterations,
		"p": parseParallelism,
	}

	for _, part := range paramParts {
		pair := strings.Split(part, "=")
		if len(pair) != hashParamPartsCount {
			return nil, errors.New("invalid hash parameters")
		}

		key, value := pair[0], pair[1]

		handler, exists := paramHandlers[key]
		if !exists {
			continue // Skip unknown parameters
		}

		if err := handler(value, params); err != nil {
			return nil, err
		}
	}

	return params, nil
}

// parseMemory handles the memory parameter
func parseMemory(value string, params *HashParams) error {
	memory, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid memory parameter: %w", err)
	}

	// Additional security check for reasonable memory limits
	if memory > MaxMemory {
		return fmt.Errorf("memory parameter exceeds maximum allowed: %d", memory)
	}

	params.Memory = uint32(memory)

	return nil
}

// parseIterations handles the iterations parameter
func parseIterations(value string, params *HashParams) error {
	iterations, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid iterations parameter: %w", err)
	}

	// Additional security check for reasonable iteration limits
	if iterations > MaxIterations {
		return fmt.Errorf("iterations parameter exceeds maximum allowed: %d", iterations)
	}

	params.Iterations = uint32(iterations)

	return nil
}

// parseParallelism handles the parallelism parameter
func parseParallelism(value string, params *HashParams) error {
	parallelism, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return fmt.Errorf("invalid parallelism parameter: %w", err)
	}

	params.Parallelism = uint8(parallelism)

	return nil
}
