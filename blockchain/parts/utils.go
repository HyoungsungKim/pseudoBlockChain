package parts

import (
	"strconv"
)

//IntToHex converse int64 to []byte
func IntToHex(integer int64) []byte {
	return []byte(strconv.FormatInt(integer, 16))
}

//ReverseBytes literally reverse
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
