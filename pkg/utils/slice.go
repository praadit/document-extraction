package utils

func Contains[T comparable](slice []T, search T) (found bool) {
	found = false
	for _, v := range slice {
		if v == search {
			found = true
			break
		}
	}
	return
}
func Distinct[T comparable](slice []T) []T {
	disinct := []T{}
	for _, v := range slice {
		if !Contains(disinct, v) {
			disinct = append(disinct, v)
		}
	}

	return disinct
}
