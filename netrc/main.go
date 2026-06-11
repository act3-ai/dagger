// A module to add a netrc with provided login credentials as a secret to any container

package main

import (
	"context"
	"crypto/sha256"
	"dagger/netrc/internal/dagger"
	"encoding/hex"
	"fmt"
	"strings"
)

type Netrc struct {
	// +private
	Logins []Login
}

type Login struct {
	// The remote machine name
	Machine string
	// username
	Username string
	// password/token
	Password *dagger.Secret
}

func New() *Netrc {
	return &Netrc{}
}

const netrcTmpl = "machine %s\nlogin %s\npassword %s\n"

// adds login credentials to netrc
func (m *Netrc) WithLogin(machine string, username string, password *dagger.Secret) *Netrc {
	m.Logins = append(m.Logins, Login{
		Machine:  machine,
		Username: username,
		Password: password,
	})
	return m
}

// creates a netrc as a secret using provided credentials in WithLogin()
func (m *Netrc) AsSecret(ctx context.Context) (*dagger.Secret, error) {
	if len(m.Logins) == 0 {
		return nil, fmt.Errorf("no logins provided; call WithLogin first")
	}

	var sb strings.Builder

	for _, login := range m.Logins {
		password, err := login.Password.Plaintext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read password secret for %s: %w", login.Machine, err)
		}

		// Format and append this machine's entry to our netrc string
		sb.WriteString(fmt.Sprintf(netrcTmpl, login.Machine, login.Username, password))
	}

	hash := sha256.Sum256([]byte(sb.String()))
	hashStr := hex.EncodeToString(hash[:])[:8]

	secretName := fmt.Sprintf("NETRC_FILE_%s", hashStr)
	netrcSecret := dag.SetSecret(secretName, sb.String())

	return netrcSecret, nil
}
