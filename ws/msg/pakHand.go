package msg

import (
	"encoding/json"
	"strings"
)

type handData struct {
	T       int32  `json:"t"`
	Ver     int64  `json:"ver"`
	Type    int32  `json:"type"`
	Charset string `json:"charset"`
	UName   string `json:"uname"`
}

type HandPak struct {
	Buff string
	Data handData
}

func (p *HandPak) RecvBuff(buff string) int {
	p.Buff += buff
	ph := MSG_HEAD
	pt := MSG_TAIL
	var headIdx, tailIdx int

	if len(buff) > 0 {
		headIdx = strings.Index(buff, ph)
		if headIdx != 0 {
			return HEAD_ILLEGAL
		}
		tailIdx = strings.Index(buff, pt)
		if tailIdx <= 0 {
			return HEAD_INCOMPLETE
		}

		subStr := buff[headIdx+len(ph) : tailIdx]
		if err := json.Unmarshal([]byte(subStr), &p.Data); err == nil {
			return HEAD_COMPLETE
		}
	}

	return HEAD_ILLEGAL
}
