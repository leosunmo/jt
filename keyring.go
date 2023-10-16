package jt

import (
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"golang.org/x/term"
)

const (
	// tokenKey is the key used to store the token in the keyring.
	tokenKey = "jira-pat"
)

func defaultKeyringConfig() keyring.Config {
	return keyring.Config{
		ServiceName:             "jt",
		LibSecretCollectionName: "jt",
		KWalletAppID:            "jt",
		KWalletFolder:           "jt",
		WinCredPrefix:           "jt",

		FilePasswordFunc:               passphrasePrompt,
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
	}
}

func GetToken() (string, error) {
	c := defaultKeyringConfig()
	kr, err := keyring.Open(c)
	if err != nil {
		return "", fmt.Errorf("failed to open keyring: %w", err)
	}

	item, err := kr.Get(tokenKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return SetToken(tokenKey)
		}

		return "", fmt.Errorf("failed to get token from keyring: %w", err)
	}
	return string(item.Data), nil
}

func SetToken(key string) (string, error) {
	c := defaultKeyringConfig()
	kr, err := keyring.Open(c)
	if err != nil {
		return "", fmt.Errorf("failed to open keyring: %w", err)
	}

	token, err := passphrasePrompt("Please enter your Jira personal access token")
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	err = kr.Set(keyring.Item{
		Key:   key,
		Label: "Jira Personal Access Token",
		Data:  []byte(token),
	})
	if err != nil {
		return "", fmt.Errorf("failed to save token to keyring: %w", err)
	}
	return token, nil
}

func passphrasePrompt(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(b), nil
}
