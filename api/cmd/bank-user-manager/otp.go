package main

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func readTOTP() (string, error) {
	fmt.Print("Enter Account TOTP: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}
