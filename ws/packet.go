package ws

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	HandHead = ""
	HandTail = ""
)

const (
	HEAD_MIN_LEN    = 32
	HEAD_INCOMPLETE = 0
	HEAD_COMPLETE   = 1
	HEAD_ILLEGAL    = -1
)

type Packet struct {
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5       string
	Data      []byte
}

type readBuf struct {
	Len       int32
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5Len    int32
}

type NetPacket struct {
	len  int
	pos  int
	hlen int
	pak  *Packet
}

func (p *NetPacket) Pack(buff []byte) (pak *Packet, err error) {
	if buff == nil {
		err = errors.New("空数据包")
		return
	}

	r := bytes.NewReader(buff)
	if p.len == 0 {
		buf := &readBuf{}
		if err = binary.Read(r, binary.LittleEndian, buf); err != nil {
			p.reset()
			return
		} else {
			md5Byte := make([]byte, buf.Md5Len)
			if err = binary.Read(r, binary.LittleEndian, &md5Byte); err != nil {
				p.reset()
				return
			} else {
				p.len = int(buf.Len)
				p.hlen = HEAD_MIN_LEN + len(md5Byte)
				p.pak = &Packet{
					Session:   buf.Session,
					Timestamp: buf.Timestamp,
					Id:        buf.Id,
					MsgId:     buf.MsgId,
					Md5:       string(md5Byte),
					Data:      make([]byte, int(buf.Len)-p.hlen),
				}
			}
		}
	}

	bytesAvail := len(buff) - p.hlen
	if bytesAvail > 0 || p.len == p.hlen {
		len := p.min(p.len-p.hlen-p.pos, bytesAvail)
		pakData := make([]byte, len)
		if err = binary.Read(r, binary.LittleEndian, &pakData); err != nil {
			p.reset()
			return
		}
		copy(p.pak.Data[p.pos:], pakData)

		p.pos += len
		if p.pos == p.len-p.hlen {
			pak = p.pak

			p.pak = nil
			p.len = 0
			p.pos = 0
			p.hlen = 0
			return
		}
	}
	return nil, nil
}

func (p *NetPacket) reset() {
	p.len = 0
	p.pos = 0
	p.hlen = 0
	p.pak = nil
}

func (p *NetPacket) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type HandData struct {
	T       int32  `json:"t"`
	Ver     int64  `json:"ver"`
	Type    int32  `json:"type"`
	Charset string `json:"charset"`
	UName   string `json:"uname"`
}

type HandPacket struct {
	Buff string
	Data HandData
}

func (p *HandPacket) RecvBuff(buff string) int {
	p.Buff += buff
	ph := HandHead
	pt := HandTail
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

type writeBuf struct {
	Len       int32
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5Len    int32
}

func GetPacketBuff(session int64, id uint, msgId int, msg []byte) (bs []byte, err error) {
	hash := md5.Sum(msg)
	md := hex.EncodeToString(hash[:])
	buf := &writeBuf{
		Session:   session,
		Timestamp: time.Now().UnixMilli(),
		Md5Len:    int32(len(md)),
		Id:        uint32(id),
		MsgId:     int32(msgId),
	}
	buf.Len = int32(HEAD_MIN_LEN) + buf.Md5Len + int32(len(msg))

	b := bytes.Buffer{}
	if err = binary.Write(&b, binary.LittleEndian, buf); err != nil {
		return
	}
	if err = binary.Write(&b, binary.LittleEndian, []byte(md)); err != nil {
		return
	}
	if err = binary.Write(&b, binary.LittleEndian, msg); err != nil {
		return
	}
	bs = b.Bytes()
	return
}
