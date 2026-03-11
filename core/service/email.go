package service

import (
	"fmt"
	"unicode"

	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// LoginEmailPassword logs in a user with an email and password by comparing
// the password with the hashed password in the database for that email.
// If the entity or email auth is not found, an error is returned.
// If the password is invalid, an error is returned.
// If the login is successful, the entity is returned.
func LoginEmailPassword(email string, password string) (model.Entity, error) {
	entity, _ := GetEntityByEmail(email)
	if entity.ID == "" {
		logger.SugarLogger.Errorf("No entity found for email %s", email)
		return model.Entity{}, fmt.Errorf("no entity found for email %s", email)
	}
	auth, _ := GetEmailAuthForEntity(entity.ID)
	if auth.EntityID == "" {
		logger.SugarLogger.Errorf("No email auth found for entity %s", entity.ID)
		return model.Entity{}, fmt.Errorf("no email auth found for entity %s", entity.ID)
	}
	err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password))
	if err != nil {
		logger.SugarLogger.Errorf("Invalid password for email %s: %v", email, err)
		return model.Entity{}, fmt.Errorf("invalid password for email %s: %v", email, err)
	}
	return entity, nil
}

// RegisterEmailPassword registers a new user with an email and password by
// hashing the password and creating a new entity and email auth for that entity.
// If the password is invalid, an error is returned.
// If the registration is successful, the entity is returned.
// Note that this function will overwrite the email auth for the entity if it already exists.
func RegisterEmailPassword(email string, password string) (model.Entity, error) {
	if err := ValidatePassword(password); err != nil {
		logger.SugarLogger.Errorf("Invalid password: %v", err)
		return model.Entity{}, err
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to hash password: %v", err)
		return model.Entity{}, err
	}
	entity, _ := GetEntityByEmail(email)
	if entity.ID == "" {
		logger.SugarLogger.Errorf("No entity found for email %s", email)
		return model.Entity{}, fmt.Errorf("no entity found for email %s, please register first", email)
	}
	auth, err := CreateEmailAuthForEntity(entity.ID, email, hashedPassword)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to add email auth for entity %s: %v", entity.ID, err)
		return model.Entity{}, err
	}
	PopulateEntity(&entity)
	logger.SugarLogger.Infof("Added email auth for entity %s: %v", entity.ID, auth)
	return entity, nil
}

// HashPassword hashes a password using bcrypt
// Returns the hashed password if successful, otherwise an error
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidatePassword checks if a password follows some basic rules.
// Returns an error if the password is invalid.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 64 {
		return fmt.Errorf("password must be at most 64 characters")
	}
	hasNumber := false
	hasCapital := false
	for _, char := range password {
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if unicode.IsUpper(char) {
			hasCapital = true
		}
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasCapital {
		return fmt.Errorf("password must contain at least one capital letter")
	}
	return nil
}
