package util

import (
	"errors"
	"net/mail"
	"strings"
)

// NormalizePhoneE164 cleans a raw phone string and returns the E.164
// form (e.g. +911234567890). v1 assumes India (+91) when no country
// code is given.

// TODO: swap in google/libphonenumber for broader
// country support when the product expands beyond IN.
func NormalizePhoneE164(raw string) (string, error) {
	s := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '-', '(', ')', '.', '\t':
			return -1
		}
		return r
	}, raw)
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty phone")
	}
	if strings.HasPrefix(s, "+") {
		if err := allDigits(s[1:]); err != nil {
			return "", err
		}
		if len(s) < 8 || len(s) > 16 {
			return "", errors.New("phone length invalid")
		}
		return s, nil
	}
	s = strings.TrimLeft(s, "0")
	if err := allDigits(s); err != nil {
		return "", err
	}
	switch {
	case len(s) == 10:
		return "+91" + s, nil
	case len(s) == 12 && strings.HasPrefix(s, "91"):
		return "+" + s, nil
	case len(s) >= 8 && len(s) <= 15:
		return "+" + s, nil
	}
	return "", errors.New("phone length invalid")
}

func allDigits(s string) error {
	for _, r := range s {
		if r < '0' || r > '9' {
			return errors.New("phone has non-digit")
		}
	}
	return nil
}

// NormalizeEmail validates and lowercase-canonicalizes an email. Returns
// the normalized address or an error. Empty input returns ("", nil) —
// email is optional in the lead pipeline.
func NormalizeEmail(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(addr.Address), nil
}
