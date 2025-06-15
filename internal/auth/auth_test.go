package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "Le4st_usele55"
	hash, _ := HashPassword(password)
	err := CheckPasswordHash(hash, password)
	if err != nil {
		t.Errorf("Password does not match the hash")
		return
	}
}

func TestMakeValidateJWT(t *testing.T) {
	user, _ := uuid.Parse("60a9b112-00f4-46bb-9e33-9b4004349d62")
	tokenSecret := "JustNot4gain"
	expiresIn := time.Duration(32154334567657)

	token, err := MakeJWT(user, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("Token was not created: %v", err)
		return
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Errorf("Token was not validated: %v", err)
		return
	}

}

func TestRejectWrongSecret(t *testing.T) {
	user, _ := uuid.Parse("60a9b112-00f4-46bb-9e33-9b4004349d62")
	tokenSecret := "JustNot4gain"
	expiresIn := time.Duration(32154334567657)

	token, err := MakeJWT(user, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("Token was not created: %v", err)
		return
	}

	wrongSecret := "Habarubu!"

	_, err = ValidateJWT(token, wrongSecret)
	if err.Error() != "token signature is invalid: signature is invalid" {
		t.Fatal("Token with wrong secret was validated!")
	}
}

func TestTokenTimeout(t *testing.T) {
	user, _ := uuid.Parse("60a9b112-00f4-46bb-9e33-9b4004349d62")
	tokenSecret := "JustNot4gain"
	expiresIn := time.Duration(1)

	token, err := MakeJWT(user, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("Token was not created: %v", err)
		return
	}
	_, err = ValidateJWT(token, tokenSecret)
	//t.Fatal(err)
	if err.Error() != "token has invalid claims: token is expired" {
		t.Fatal("Timed out token was validated!")
	}
}

func TestGetBearerToken(t *testing.T) {
	user, _ := uuid.Parse("60a9b112-00f4-46bb-9e33-9b4004349d62")
	tokenSecret := "JustNot4gain"
	expiresIn := time.Duration(32154334567657)

	token, err := MakeJWT(user, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("Token was not created: %v", err)
		return
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	retrievedToken, err := GetBearerToken(header)
	if err != nil {
		t.Errorf("Token was not recovered: %v", err)
		return
	}

	if retrievedToken != token {
		t.Error("Tokens do not coincide!\n")
		t.Errorf("retrieved %v\n", retrievedToken)
		t.Errorf("initial: %v\n", token)
		return
	}

}
