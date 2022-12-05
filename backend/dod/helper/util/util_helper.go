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

// Removes all duplicates form the input list, only returning the unique items in the list
// for example
// []string{"string1","string1","string1","string2"}
// Will be returned as:
// []string{"string1","string2"}
func FilterUniqueStrings(list *[]string) *[]string {
	var uniqueList []string
	for _, e := range *list {
		if IsStringUnique(&uniqueList, e) {
			uniqueList = append(uniqueList, e)
		}
	}
	return &uniqueList
}
