package util

// Sub returns all the element in arr1 but not in arr2
func Sub(arr1 []string, arr2 []string) []string {
	result := make([]string, 0)
	for _, s := range arr1 {
		exist := false
		for _, s2 := range arr2 {
			if s == s2 {
				exist = true
			}
		}
		if !exist {
			result = append(result, s)
		}
	}
	return result
}

// IsSameStringArray returns true if arr1 and arr2 has exactly same elements without order care
func IsSameStringArray(arr1 []string, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for _, s := range arr1 {
		exist := false
		for _, s2 := range arr2 {
			if s2 == s {
				exist = true
				break
			}
		}
		if !exist {
			return false
		}
	}
	return true
}
