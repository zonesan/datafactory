package util

func Contains(arr []string, str string) bool {
	for _, ele := range arr {
		if ele == str {
			return true
		}
	}
	return false
}
