package util

// Checks if the item is unique and does not already exist in the list
func IsStringUnique(list *[]string, item string) bool {
	for _, e := range *list {
		if e == item {
			return false
		}
	}
	return true
}
