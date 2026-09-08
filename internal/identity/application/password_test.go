package application

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	encoded, err := hashPassword("correct horse battery staple", "pepper")
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(encoded, "correct horse battery staple", "pepper") {
		t.Fatal("valid password was rejected")
	}
	if verifyPassword(encoded, "wrong password", "pepper") {
		t.Fatal("invalid password was accepted")
	}
}

func TestTokenHashRoundTrip(t *testing.T) {
	token, expected, err := newToken()
	if err != nil {
		t.Fatal(err)
	}
	actual, err := tokenHash(token)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(expected) {
		t.Fatal("token hash changed")
	}
}
