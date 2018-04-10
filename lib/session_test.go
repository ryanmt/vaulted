package vaulted

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestSessionVariables(t *testing.T) {
	e := Session{
		Name:       "vault",
		Expiration: time.Now(),
		Vars: map[string]string{
			"TEST":         "TESTING",
			"ANOTHER_TEST": "TEST TEST",
		},
	}
	var expectedSet map[string]string = map[string]string{
		"ANOTHER_TEST":           "TEST TEST",
		"TEST":                   "TESTING",
		"VAULTED_ENV":            e.Name,
		"VAULTED_ENV_EXPIRATION": e.Expiration.UTC().Format(time.RFC3339),
	}
	var expectedUnset []string

	vars := e.Variables()

	if !reflect.DeepEqual(expectedSet, vars.Set) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedSet, vars.Set)
	}

	if !reflect.DeepEqual(expectedUnset, vars.Unset) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedUnset, vars.Unset)
	}
}

func TestSessionVariablesWithPermCreds(t *testing.T) {
	e := Session{
		Name:       "vault",
		Expiration: time.Now(),
		AWSCreds: &AWSCredentials{
			ID:     "an-id",
			Secret: "the-super-sekrit",
		},
		Vars: map[string]string{
			"TEST":         "TESTING",
			"ANOTHER_TEST": "TEST TEST",
		},
	}
	var expectedSet map[string]string = map[string]string{
		"ANOTHER_TEST":           "TEST TEST",
		"AWS_ACCESS_KEY_ID":      e.AWSCreds.ID,
		"AWS_SECRET_ACCESS_KEY":  e.AWSCreds.Secret,
		"TEST":                   "TESTING",
		"VAULTED_ENV":            e.Name,
		"VAULTED_ENV_EXPIRATION": e.Expiration.UTC().Format(time.RFC3339),
	}
	var expectedUnset []string = []string{
		"AWS_SECURITY_TOKEN",
		"AWS_SESSION_TOKEN",
	}

	vars := e.Variables()

	if !reflect.DeepEqual(expectedSet, vars.Set) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedSet, vars.Set)
	}

	sort.Strings(vars.Unset)
	if !reflect.DeepEqual(expectedUnset, vars.Unset) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedUnset, vars.Unset)
	}
}

func TestSessionVariablesWithTempCreds(t *testing.T) {
	e := Session{
		Name:       "vault",
		Expiration: time.Now(),
		AWSCreds: &AWSCredentials{
			ID:     "an-id",
			Secret: "the-super-sekrit",
			Token:  "my-affections",
		},
		Vars: map[string]string{
			"TEST":         "TESTING",
			"ANOTHER_TEST": "TEST TEST",
		},
	}
	var expectedSet map[string]string = map[string]string{
		"ANOTHER_TEST":           "TEST TEST",
		"AWS_ACCESS_KEY_ID":      e.AWSCreds.ID,
		"AWS_SECRET_ACCESS_KEY":  e.AWSCreds.Secret,
		"AWS_SECURITY_TOKEN":     e.AWSCreds.Token,
		"AWS_SESSION_TOKEN":      e.AWSCreds.Token,
		"TEST":                   "TESTING",
		"VAULTED_ENV":            e.Name,
		"VAULTED_ENV_EXPIRATION": e.Expiration.UTC().Format(time.RFC3339),
	}
	var expectedUnset []string

	vars := e.Variables()

	if !reflect.DeepEqual(expectedSet, vars.Set) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedSet, vars.Set)
	}

	if !reflect.DeepEqual(expectedUnset, vars.Unset) {
		t.Errorf("Expected: %#v\nGot: %#v\n", expectedUnset, vars.Unset)
	}
}

func TestSessionMerge(t *testing.T) {
	child := Session{
		Name:       "vault",
		Expiration: time.Now().Add(time.Hour),
		Vars: map[string]string{
			"TEST":         "TESTING",
			"ANOTHER_TEST": "TEST TEST",
		},
	}
	parent := Session{
		Name:       "vault2",
		Expiration: time.Now().Add(time.Minute),
		AWSCreds: &AWSCredentials{
			ID:     "an-id",
			Secret: "the-super-sekrit",
			Token:  "my-affections",
		},
		Vars: map[string]string{
			"TEST":            "FAIL",
			"ANOTHER_ANOTHER": "TEST TEST TEST",
		},
	}

	resultantSession := parent.mergeFrom(child)

	if resultantSession.Name != parent.Name+"/"+child.Name {
		t.Error("We got the wrong name!!")
	}

	if resultantSession.AWSCreds.ID != parent.AWSCreds.ID {
		t.Error("AWS Creds didn't merge as expected.")
	}
	if resultantSession.AWSCreds.Secret != parent.AWSCreds.Secret {
		t.Error("AWS Creds didn't merge as expected.")
	}
	if resultantSession.AWSCreds.Token != parent.AWSCreds.Token {
		t.Error("AWS Creds didn't merge as expected.")
	}

	if resultantSession.Vars["ANOTHER_ANOTHER"] != parent.Vars["ANOTHER_ANOTHER"] {
		t.Error("Vars didn't merge correctly.")
	}
	if resultantSession.Vars["TEST"] != child.Vars["TEST"] {
		t.Error("Vars didn't merge correctly.")
	}
	if resultantSession.Vars["TEST"] == parent.Vars["TEST"] {
		t.Error("Vars didn't merge correctly.")
	}
}
