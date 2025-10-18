package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Password string `db:"password" json:"password"`
	decoded  string `db:"-" json:"-"`
}

func (p *Password) Set(oldPassword, newPassword string) error {
	if !p.Verify(oldPassword) {
		return errors.New("old password does not match")
	}

	encoded, err := p.EncodePassword(newPassword)
	if err != nil {
		return err
	}

	p.Password = encoded
	return nil
}

func (p *Password) EncodePassword(encode string) (encoded string, encondingError error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(encode), bcrypt.DefaultCost)

	return string(bytes), err
}

func (p *Password) Verify(verify string) bool {
	if p.Password == "" {
		return true
	}

	err := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(verify))
	if err == nil {
		p.decoded = verify
	}

	return err == nil
}

func (p *Password) Encrypt(encode string) (encrypted string, encryptionError error) {
	plainText := []byte(encode)
	key := p.deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	encryptedText := gcm.Seal(nonce, nonce, plainText, nil)

	return string(encryptedText), nil
}

func (p *Password) Decrypt(cipherText string) (plainText string, err error) {
	key := p.deriveKey()
	encryptedText := []byte(cipherText)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(encryptedText) < gcm.NonceSize() {
		return "", errors.New("Malformed encripted string")
	}

	plainTextBytes, err := gcm.Open(nil,
		encryptedText[:gcm.NonceSize()],
		encryptedText[gcm.NonceSize():],
		nil,
	)

	return string(plainTextBytes), err
}

func (p *Password) deriveKey() []byte {
	hash := sha256.Sum256([]byte(p.decoded))
	return hash[:]
}
