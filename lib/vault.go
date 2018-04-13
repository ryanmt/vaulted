package vaulted

import (
	"errors"
	"strings"
	"time"
)

const (
	DefaultSessionName = "VaultedSession"
)

var STSDurationDefault = time.Hour

var (
	ErrInvalidCommand = errors.New("Invalid command")
	ErrNoTokenEntered = errors.New("Could not get MFA code")
)

type Vault struct {
	Duration  time.Duration     `json:"duration,omitempty"`
	AWSKey    *AWSKey           `json:"aws_key,omitempty"`
	Vars      map[string]string `json:"vars,omitempty"`
	SSHKeys   map[string]string `json:"ssh_keys,omitempty"`
	SubVaults map[string]*Vault `json:"subvaults,omitempty"`
}

func (v *Vault) CreateSession(name string) (*Session, error) {
	return v.createSessionWrapper(name, func(duration time.Duration) (*AWSCredentials, error) {
		return v.AWSKey.GetAWSCredentials(duration)
	})
}

func (v *Vault) CreateSessionWithMFA(name, mfaToken string) (*Session, error) {
	return v.createSessionWrapper(name, func(duration time.Duration) (*AWSCredentials, error) {
		return v.AWSKey.GetAWSCredentialsWithMFA(mfaToken, duration)
	})
}

func (v *Vault) createSessionWrapper(name string, credsFunc func(duration time.Duration) (*AWSCredentials, error)) (*Session, error) {
	names := strings.Split(name, "/")
	session, err := v.createSession(names, 1, credsFunc)
	if err == nil {
		session.Name = name
	}
	return session, err
}

func (v *Vault) createSession(path []string, cursor int, credsFunc func(duration time.Duration) (*AWSCredentials, error)) (*Session, error) {
	baseSession, err := v.constructSession(path[0:cursor], credsFunc)
	if err != nil {
		return nil, err
	}

	if len(path) > cursor {
		subVaultPath, remainingPath := path[0:cursor], path[cursor:]
		subVaultName := path[cursor]
		vault := v.mergeFrom(v.SubVaults[subVaultName])

		var newSession *Session
		if len(remainingPath) > 0 {
			newSession, err = vault.createSession(path, cursor+1, credsFunc)
		} else {
			newSession, err = vault.constructSession(subVaultPath, credsFunc)
		}
		if err != nil {
			return nil, err
		}

		baseSession.SubSessions[subVaultName] = newSession
	}

	return baseSession, nil
}

func (v *Vault) constructSession(path []string, credsFunc func(duration time.Duration) (*AWSCredentials, error)) (*Session, error) {
	var duration time.Duration
	if v.Duration == 0 {
		duration = STSDurationDefault
	} else {
		duration = v.Duration
	}

	// For each vault in VaultSet...
	session := &Session{
		Name:        joinNames(path),
		Vars:        make(map[string]string),
		SubSessions: make(map[string]*Session),
	}

	// copy the vault vars to the session
	for key, value := range v.Vars {
		session.Vars[key] = value
	}

	// copy the vault ssh keys to the session
	if len(v.SSHKeys) > 0 {
		session.SSHKeys = make(map[string]string)
		for key, value := range v.SSHKeys {
			session.SSHKeys[key] = value
		}
	}

	if v.AWSKey.Valid() {
		var err error
		session.AWSCreds, err = credsFunc(duration)
		if err != nil {
			return nil, err
		}
		session.Role = v.AWSKey.Role
	}

	// now that the session is generated, set the expiration
	session.Expiration = time.Now().Add(duration).Truncate(time.Second)
	return session, nil
}

func (v *Vault) mergeFrom(child *Vault) *Vault {
	newVars := stringMapMerge(v.Vars, child.Vars)
	newSSHKeys := stringMapMerge(v.SSHKeys, child.SSHKeys)
	newDuration := v.Duration
	if newDuration > child.Duration {
		newDuration = child.Duration
	}

	newVault := &Vault{
		Duration: newDuration,
		Vars:     newVars,
		SSHKeys:  newSSHKeys,
	}

	if v.AWSKey != nil {
		newVault.AWSKey = v.AWSKey
	}
	if child.AWSKey != nil {
		newVault.AWSKey = child.AWSKey
	}

	return newVault
}

func (v *Vault) subVaultNames(baseName string) []string {
	outputs := []string{baseName}
	for svName := range v.SubVaults {
		prefix := baseName + "/" + svName
		outputs = append(outputs, v.SubVaults[svName].subVaultNames(prefix)...)
	}
	return outputs
}

// CombineVaults merge each vault to produce a representation of that vault
func (v *Vault) CombineVaults(vaults []*Vault) *Vault {
	resultant, children := vaults[0], vaults[1:]
	for _, child := range children {
		resultant = resultant.mergeFrom(child)
	}
	return resultant
}
