package auth

import (
"github.com/alexedwards/argon2id"
"github.com/golang-jwt/golang-jwt"
)

fucn MakeJWT(userID, uuid.UUID, tokensecret string, expieresIn time.Duration) (string, error) {
	
}

func HashPassword(password string) (string, error) {

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams) 
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}
