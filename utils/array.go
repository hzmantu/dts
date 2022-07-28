package utils

func InStringArray(in string, lists []string) bool {
	for _, core := range lists {
		if core == in {
			return true
		}
	}
	return false
}

func InIntArray(in int, lists []int) bool {
	for _, core := range lists {
		if core == in {
			return true
		}
	}
	return false
}
