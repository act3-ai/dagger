// A module to add a netrc with provided login credentials as a secret to any container

package main

import (
	"context"
	"dagger/netrc/internal/dagger"
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

	netrcSecret := dag.SetSecret("netrc-file", sb.String())

	return netrcSecret, nil
}
