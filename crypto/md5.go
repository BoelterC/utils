package crypto

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func MD5(str string, lower bool) string {
	data := []byte(str)
	result := fmt.Sprintf("%x", md5.Sum(data))
	if !lower {
		result = strings.ToUpper(result)
	}
	return result
}
