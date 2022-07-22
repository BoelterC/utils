package crypto

import (
	"fmt"
	"time"
)

func GenKeyWord(key string, secs int64, maxlen int) string {
	if secs == 0 {
		secs = 30
	}
	if maxlen <= 0 {
		maxlen = 6
	}

	key = key + fmt.Sprintf("%d", time.Now().Unix()/secs)
	md5 := MD5(key, false)
	if maxlen >= len(md5) {
		return md5
	}

	return md5[:maxlen]
}
