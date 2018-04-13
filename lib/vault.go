package vaulted

import (
	"errors"
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
	return v.createSession(name, func(duration time.Duration) (*AWSCredentials, error) {
		return v.AWSKey.GetAWSCredentials(duration)
	})
}

func (v *Vault) CreateSessionWithMFA(name, mfaToken string) (*Session, error) {
	return v.createSession(name, func(duration time.Duration) (*AWSCredentials, error) {
		return v.AWSKey.GetAWSCredentialsWithMFA(mfaToken, duration)
	})
}

func (v *Vault) createSession(name string, credsFunc func(duration time.Duration) (*AWSCredentials, error)) (*Session, error) {
	baseName, names := splitNames(name)
	baseSession, err := v.constructSession(baseName, credsFunc)
	if err != nil {
		return nil, err
	}

	if len(names) > 0 {
		subVaultName, nextNames := names[0], names[1:]
		vault := v.mergeFrom(v.SubVaults[subVaultName])

		var newSession *Session
		if len(nextNames) > 0 {
			newSession, err = vault.createSession(joinNames(nextNames), credsFunc)
		} else {
			newSession, err = vault.constructSession(subVaultName, credsFunc)
		}
		if err != nil {
			return nil, err
		}

		baseSession.SubSessions[subVaultName] = newSession
	}

	return baseSession, nil
}

func (v *Vault) constructSession(name string, credsFunc func(duration time.Duration) (*AWSCredentials, error)) (*Session, error) {
	var duration time.Duration
	if v.Duration == 0 {
		duration = STSDurationDefault
	} else {
		duration = v.Duration
	}

	// For each vault in VaultSet...
	session := &Session{
		Name:        name,
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

// This function is returning too much, not the path, but all of the paths
// func (v *Vault) getVaultSet(name string) ([]*Vault, error) {
// _, names := splitNames(name)
// var vaultSet []*Vault

// if len(names) > 0 {
// var vaults
// vaults, err := v.crawlVaultPath(names)
// if err == ErrSubvaultDoesNotExist {
// return nil, fmt.Errorf("Subvault %s not found", name)
// }
// if err != nil {
// return nil, err
// }
// vaultSet = append([]*Vault{v}, vaults...)

// }
// return vaultSet, nil
// }

// CombineVaults merge each vault to produce a representation of that vault
func (v *Vault) CombineVaults(vaults []*Vault) *Vault {
	resultant, children := vaults[0], vaults[1:]
	for _, child := range children {
		resultant = resultant.mergeFrom(child)
	}
	return resultant
}

func (v *Vault) crawlVaultPath(names []string, vaultHandler func(vault *Vault, path string) error) ([]*Vault, error) {
	name, nextNames := names[0], names[1:]
	nextVault := v.SubVaults[name]
	if nextVault != nil {
		if len(nextNames) > 0 {
			nextSet, err := nextVault.crawlVaultPath(nextNames, vaultHandler)
			if err != nil {
				return nil, err
			}
			return append([]*Vault{nextVault}, nextSet...), nil
		}
		return []*Vault{nextVault}, nil
	}
	return nil, ErrSubvaultDoesNotExist
}
