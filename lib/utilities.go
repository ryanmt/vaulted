package vaulted

func stringMapMerge(parent map[string]string, child map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range parent {
		output[k] = v
	}
	// Add all values from child vault and overwrite
	for k, v := range child {
		output[k] = v
	}
	return output
}
