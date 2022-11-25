package msg

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type readBuf struct {
	Len       int32
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5Len    int32
}

type MsgPak struct {
	len  int
	pos  int
	hlen int
	pak  *Packet
}

func (p *MsgPak) Pack(buff []byte) (pak *Packet, err error) {
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

func (p *MsgPak) reset() {
	p.len = 0
	p.pos = 0
	p.hlen = 0
	p.pak = nil
}

func (p *MsgPak) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
