package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"
)

type License struct {
	Key            string
	User           string
	Email          string
	Product        string
	Expiration     time.Time
	Status         string
	originalLicKey string
}

type LicenseManager struct {
	licenses map[string]License
}

func EncryptLicenseKey(key string, secret []byte) (string, error) {
	block, err := aes.NewCipher((secret))
	if err != nil {
		return "", err
	}
	padding := block.BlockSize() - len(key)%block.BlockSize()
	paddedKey := append([]byte(key), make([]byte, padding)...)
	// Add random initialization vector
	iv := make([]byte, block.BlockSize())
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(paddedKey, paddedKey)
	return fmt.Sprintf("%x", iv) + fmt.Sprintf("%x", paddedKey), nil
}

func DecryptLicenseKey(encryptedKey string, secret []byte) (string, error) {
	data, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}
	// Split the data into IV and encrypted license key
	iv := data[:aes.BlockSize]
	encryptedKeyData := data[aes.BlockSize:]
	// Build  AES block with cipher with the secret key
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", errors.ErrUnsupported
	}
	// Decrypt it
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encryptedKeyData, encryptedKeyData)

	// Return the decrypted key as a string
	return string(encryptedKeyData), nil
}

func GenerateLicenseKey(user string) string {
	// Get the current timestamp
	timestamp := time.Now().Unix()

	// Combine user information and timestamp to create a seed for the license key
	data := fmt.Sprintf("%s:%d", user, timestamp)

	// Use SHA-256 to hash the seed and create a unique license key
	hash := sha256.New()
	hash.Write([]byte(data))
	licenseKey := hex.EncodeToString(hash.Sum(nil))

	return string(licenseKey)
}

func (lm *LicenseManager) CreateLicense(user, product, email string, expiration time.Time, secret []byte) (string, error) {
	licenseKey := GenerateLicenseKey(user)
	encryptedKey, err := EncryptLicenseKey(licenseKey, secret)
	if err != nil {
		return "", err
	}
	license := License{
		Key:            encryptedKey,
		User:           user,
		Email:          email,
		Product:        product,
		Expiration:     expiration,
		Status:         "Active",
		originalLicKey: licenseKey,
	}

	// Store the encrypted license key in the map
	lm.licenses[encryptedKey] = license

	return encryptedKey, nil

}
func warning(msg string) {
	log.Println("WARNING:", msg)
}

func fatal(msg string) {
	log.Fatalln("FATAL:", msg)
}
func NewLicenseManager() *LicenseManager {
	return &LicenseManager{
		licenses: make(map[string]License),
	}
}

func (lm *LicenseManager) ValidateLicense(encryptedKey string, secret []byte) (*License, error) {
	// Find the license using the encrypted key
	license, exists := lm.licenses[encryptedKey]
	if !exists {
		return nil, errors.New("license key not found")
	}

	// Decrypt the license key using the secret
	decryptedKey, err := DecryptLicenseKey(license.Key, secret)
	if err != nil {
		return nil, err
	}

	originalKey := license.originalLicKey
	compare_length := len(originalKey)
	if len(decryptedKey) != compare_length {
		if len(originalKey) > len(decryptedKey) {
			compare_length = len(decryptedKey)
		}
	}
	if originalKey[0:compare_length] != decryptedKey[0:compare_length] {
		warning("Invalid license key")
		return nil, errors.New("invalid license key")
	}
	// Check if the license is expired
	if license.Expiration.Before(time.Now()) {
		license.Status = "Expired"
		lm.licenses[encryptedKey] = license
		return nil, errors.New("license has expired")
	}

	// Check if the license is deactivated
	if license.Status == "Deactivated" {
		return nil, errors.New("license is deactivated")
	}

	// License is valid
	return &license, nil
}
