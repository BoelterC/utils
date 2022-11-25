package msg

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"time"
)

type writeBuf struct {
	Len       int32
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5Len    int32
	Md5       [16]byte
}

type Packet struct {
	Session   int64
	Timestamp int64
	Id        uint32
	MsgId     int32
	Md5       string
	Data      []byte
}

func MakePacket(session int64, id uint, msgId int, msg []byte) (bs []byte, err error) {
	md := md5.Sum(msg)
	buf := &writeBuf{
		Session:   session,
		Timestamp: time.Now().UnixNano(),
		Md5Len:    int32(len(md)),
		Md5:       md,
		Id:        uint32(id),
		MsgId:     int32(msgId),
	}
	buf.Len = int32(HEAD_MIN_LEN) + buf.Md5Len + int32(len(msg))

	b := bytes.Buffer{}
	if err = binary.Write(&b, binary.LittleEndian, buf); err != nil {
		return
	}
	if err = binary.Write(&b, binary.LittleEndian, msg); err != nil {
		return
	}
	bs = b.Bytes()
	return
}
