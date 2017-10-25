package vaulted

func stringMapMerge(first map[string]string, second map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range second {
		output[k] = v
	}
	// Add all values from first vault and overwrite
	for k, v := range first {
		output[k] = v
	}
	return output
}
