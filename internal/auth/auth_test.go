package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "someSecret"
	wrongSecret := "wrongsecret"
	validToken, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT failed to create valid token %v", err)
	}
	expieredToken, err := MakeJWT(userID, secret, -time.Second)
	if err != nil {
		t.Fatalf("MakeJWT failed to create expiered token %v", err)
	}
	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   bool
		wantMatch bool
	}{
		{
			name:      "valid token, valid experation",
			token:     validToken,
			secret:    secret,
			wantErr:   false,
			wantMatch: true,
		},
		{
			name:      "invalid secret",
			token:     validToken,
			secret:    wrongSecret,
			wantErr:   true,
			wantMatch: false,
		},
		{
			name:      "Expiered token",
			token:     expieredToken,
			secret:    secret,
			wantErr:   true,
			wantMatch: false,
		},
		{
			name:      "Malformed token",
			token:     "no-jwt-here",
			secret:    secret,
			wantErr:   true,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := ValidateJWT(tt.token, tt.secret)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("did not expect error, got: %v", err)
			}

			if tt.wantMatch && gotID != userID {
				t.Errorf("expected userID %v, got %v", userID, gotID)
			}

			if !tt.wantMatch && gotID == userID {
				t.Errorf("expected no match but got the same userID")
			}
		})
	}

}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	stipedToken := "abc123"
	headers.Set("Authorization", "Bearer abc123")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken() had an error %v", err)
	}
	if token != stipedToken {
		t.Errorf("expected %q, got %q", stipedToken, token)
	}
}
