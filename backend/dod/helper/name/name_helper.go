package name

import "strconv"

func Network(prefix string, id uint) string {
	return prefix + strconv.Itoa(int(id))
}
