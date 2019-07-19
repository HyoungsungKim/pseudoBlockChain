package parts

import (
	"strconv"
)

//IntToHex converse int64 to []byte
func IntToHex(integer int64) []byte {
	return []byte(strconv.FormatInt(integer, 16))
}
