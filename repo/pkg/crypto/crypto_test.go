package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKey = "0123456789abcdef0123456789abcdef"

func TestNewFieldEncryptor_ValidKey(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)
	assert.NotNil(t, enc)
}

func TestNewFieldEncryptor_RejectsShortKey(t *testing.T) {
	_, err := NewFieldEncryptor("short")
	assert.Error(t, err)
}

func TestNewFieldEncryptor_RejectsLongKey(t *testing.T) {
	_, err := NewFieldEncryptor("0123456789abcdef0123456789abcdef_too_long")
	assert.Error(t, err)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)

	plaintext := "user@example.com"
	ct, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, ct)
	assert.NotEqual(t, plaintext, ct, "ciphertext must not equal plaintext")

	pt, err := enc.Decrypt(ct)
	require.NoError(t, err)
	assert.Equal(t, plaintext, pt)
}

func TestEncrypt_ProducesDistinctCiphertextsForSamePlaintext(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)

	ct1, err := enc.Encrypt("same")
	require.NoError(t, err)
	ct2, err := enc.Encrypt("same")
	require.NoError(t, err)

	// GCM uses a random nonce, so the same plaintext must yield different
	// ciphertexts — this is a non-trivial security invariant worth asserting.
	assert.NotEqual(t, ct1, ct2)
}

func TestDecrypt_RejectsCorruptedCiphertext(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)

	ct, err := enc.Encrypt("secret")
	require.NoError(t, err)

	// Flip a byte at the end of the base64 string to corrupt the GCM tag.
	corrupted := ct[:len(ct)-1] + "X"
	_, err = enc.Decrypt(corrupted)
	assert.Error(t, err)
}

func TestDecrypt_RejectsInvalidBase64(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)

	_, err = enc.Decrypt("!!!not-base64!!!")
	assert.Error(t, err)
}

func TestDecrypt_RejectsTooShortCiphertext(t *testing.T) {
	enc, err := NewFieldEncryptor(testKey)
	require.NoError(t, err)

	// Valid base64 but too short to contain a nonce.
	_, err = enc.Decrypt("YWJj") // "abc"
	assert.Error(t, err)
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"user@example.com", "u***@example.com"},
		{"a@example.com", "a***@example.com"},
		{"abcdef@example.org", "a***@example.org"},
		{"no-at-sign", "***"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, MaskEmail(c.in), "input=%s", c.in)
	}
}
