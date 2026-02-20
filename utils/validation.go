// validation.go contains application logic.
package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var phoneRe = regexp.MustCompile(`^\+?[0-9]{8,15}$`)
var pincodeRe = regexp.MustCompile(`^[0-9]{4,10}$`)

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return false
	}
	return phoneRe.MatchString(phone)
}

func ValidatePincode(pincode string) bool {
	pincode = strings.TrimSpace(pincode)
	if pincode == "" {
		return false
	}
	return pincodeRe.MatchString(pincode)
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password too short")
	}
	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber || !hasSymbol {
		return errors.New("password must include upper, lower, number, and symbol")
	}
	return nil
}

