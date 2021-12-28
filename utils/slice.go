package utils

func Contains(arr []string, s string) bool {
	for i := range arr {
		if arr[i] == s {
			return true
		}
	}

	return false
}
