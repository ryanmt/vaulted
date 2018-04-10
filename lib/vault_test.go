package vaulted

import (
	"testing"
	"time"
)

var creds AWSCredentials
var key *AWSKey
var parentCreds AWSCredentials
var parentKey *AWSKey
var child Vault
var parent Vault

func init() {
	creds = AWSCredentials{
		ID:     "an-id",
		Secret: "the-super-sekrit",
		Token:  "my-affections",
	}
	key = &AWSKey{
		AWSCredentials: creds,
		MFA:            "token",
		Role:           "role",
		ForgoTempCredGeneration: false,
	}
	parentCreds = AWSCredentials{
		ID:     "parent-id",
		Secret: "parent-sekrit",
		Token:  "parent-token",
	}
	parentKey = &AWSKey{
		AWSCredentials: parentCreds,
		MFA:            "parent_token",
		Role:           "parent_role",
		ForgoTempCredGeneration: true,
	}
	child = Vault{
		AWSKey:   key,
		Duration: time.Hour,
		SSHKeys: map[string]string{
			"KEY1": "key1",
		},
		Vars: map[string]string{
			"TEST": "TESTING",
		},
	}

	childVaultName := "child"
	subVaults := make(map[string]*Vault)
	subVaults[childVaultName] = &child

	parent = Vault{
		AWSKey:   parentKey,
		Duration: time.Minute,
		SSHKeys: map[string]string{
			"KEY1": "fail",
			"KEY2": "key2",
		},
		Vars: map[string]string{
			"TEST":         "FAIL",
			"ANOTHER_TEST": "TEST TEST TEST",
		},
		SubVaults: subVaults,
	}
}

func TestVaultMerge(t *testing.T) {
	resultVault := parent.mergeFrom(child)

	if resultVault.Duration != time.Minute {
		t.Error("Duration should be the shortest value")
	}

	if resultVault.AWSKey.MFA != child.AWSKey.MFA {
		t.Error("AWS Key didn't merge as expected.")
	}
	if resultVault.AWSKey.Role != child.AWSKey.Role {
		t.Error("AWS Key didn't merge as expected.")
	}
	if resultVault.AWSKey.ForgoTempCredGeneration != child.AWSKey.ForgoTempCredGeneration {
		t.Error("AWS Key didn't merge as expected.")
	}

	if resultVault.AWSKey.AWSCredentials.ID != child.AWSKey.AWSCredentials.ID {
		t.Error("AWS Creds didn't merge as expected.")
	}
	if resultVault.AWSKey.AWSCredentials.Secret != child.AWSKey.AWSCredentials.Secret {
		t.Error("AWS Creds didn't merge as expected.")
	}
	if resultVault.AWSKey.AWSCredentials.Token != child.AWSKey.AWSCredentials.Token {
		t.Error("AWS Creds didn't merge as expected.")
	}

	if resultVault.Vars["ANOTHER_TEST"] != parent.Vars["ANOTHER_TEST"] {
		t.Log(resultVault.Vars)
		t.Log(parent.Vars)
		t.Error("Vars didn't merge correctly.")
	}
	if resultVault.Vars["TEST"] == parent.Vars["TEST"] {
		t.Error("Vars didn't merge correctly.")
	}
	if resultVault.Vars["TEST"] != child.Vars["TEST"] {
		t.Error("Vars didn't merge correctly.")
	}

	if resultVault.SSHKeys["KEY1"] != child.SSHKeys["KEY1"] {
		t.Error("SSHKeys didn't merge correctly.")
	}
	if resultVault.SSHKeys["KEY1"] == parent.SSHKeys["KEY1"] {
		t.Error("SSHKeys didn't merge correctly.")
	}
	if resultVault.SSHKeys["KEY2"] != parent.SSHKeys["KEY2"] {
		t.Error("SSHKeys didn't merge correctly.")
	}

	// Confirm we aren't shallow copying
	if resultVault.AWSKey.AWSCredentials.ID != "an-id" {
		t.Error("Values not as expected")
	}
	if parent.AWSKey.AWSCredentials.ID != "parent-id" {
		t.Error("It appears that the original altered")
	}
	if resultVault.AWSKey.Role != "role" {
		t.Error("Value not as expected")
	}
	if parent.AWSKey.Role != "parent_role" {
		t.Error("It appears that the original altered")
	}
	if child.AWSKey.Role != "role" {
		t.Error("It appears that the original altered")
	}
	if resultVault.AWSKey.AWSCredentials.ID != "an-id" {
		t.Error("Value not as expected")
	}

	// Without any AWSKey
	resultVault = child.mergeFrom(child)
	if resultVault.AWSKey != child.AWSKey {
		t.Error("AWS Key didn't merge as expected.")
	}
}

func TestVaultsubVaultNames(t *testing.T) {
	result := parent.subVaultNames()
	expectation := []string{"parent", "parent/child"}
	if result[0] != expectation[0] {
		t.Error("parent name not properly set")
	}
	if result[1] != expectation[1] {
		t.Error("subvault name not properly set")
	}
}
