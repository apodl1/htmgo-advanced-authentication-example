package user

import (
  "crypto/rand"
  "crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func HashPassword(password string) (string, error) {
	// Recommended parameters for Argon2id (RFC 9106, retrieved 2025)
	p := &argonParams{
			memory:      64 * 1024, // 64MB
			iterations:  3,
			parallelism: 2, // Default value, adjust per core count if needed
			saltLength:  16,
			keyLength:   32,
		}

	salt := make([]byte, p.saltLength)
		if _, err := rand.Read(salt); err != nil {
			return "", err
		}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedPassword := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64Salt,
		b64Hash,
	)

	return string(encodedPassword), nil
}

func PasswordMatches(password string, hashedPassword string) bool {
	  // Parse the parts (see above) of the database entry
		parts := strings.Split(hashedPassword, "$")
		if len(parts) != 6 {
			return false
		}

		// Extract the argon2 parameters
		var p argonParams
		_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
		if err != nil {
			return false
		}

		// Take the salt and decode to bytes
		salt, err := base64.RawStdEncoding.DecodeString(parts[4])
		if err != nil {
			return false
		}

		// Take the hashed password and decode to bytes
		decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false
		}

		// Re-hash the input password with the same parameters
		comparisonHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, uint32(len(decodedHash)))

		// Use ConstantTimeCompare to prevent timing attacks
		return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1
}
