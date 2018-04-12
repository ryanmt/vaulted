package vaulted

import "strings"

func stringMapMerge(base map[string]string, overrides map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range base {
		output[k] = v
	}
	// Add all values from child vault and overwrite
	for k, v := range overrides {
		output[k] = v
	}
	return output
}

// splitNames Return the vault and subvaults from a given name, maybe an argument for moving this to command
func splitNames(name string) (string, []string) {
	names := strings.Split(name, "/")
	return names[0], names[1:]
}
