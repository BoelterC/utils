package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"nhooyr.io/websocket"
)

const (
	st_close = iota
	st_init
	st_open
)

type client struct {
	st        int
	session   int64
	rcvCh     chan error
	msgs      chan []byte
	txts      chan string
	pbPak     chan *Packet
	closeSlow func()
	handPak   *HandPacket
	netPak    *NetPacket
}

func NewClient(uid int64, msgBuff int, closeFunc func()) *client {
	return &client{
		st:        st_init,
		session:   uid,
		rcvCh:     make(chan error),
		msgs:      make(chan []byte, msgBuff),
		txts:      make(chan string, msgBuff),
		pbPak:     make(chan *Packet, msgBuff),
		closeSlow: closeFunc,
		handPak:   &HandPacket{},
		netPak:    &NetPacket{},
	}
}

func (cl *client) Read(ctx context.Context, c *websocket.Conn) error {
	_, bytes, err := c.Read(ctx)
	if err != nil {
		return err
	}

	if cl.st == st_init {
		result := cl.handPak.RecvBuff(string(bytes))
		if result == HEAD_ILLEGAL {
			cl.st = st_close
			return errors.New("握手失败！非法的消息格式")
		} else if result == HEAD_COMPLETE {
			cl.st = st_open
			bytes, _ := json.Marshal(map[string]interface{}{"msg": "hand shake success!", "session": cl.session})
			cl.txts <- fmt.Sprintf("%s%s%s", HandHead, bytes, HandTail)
		}
	} else if cl.st == st_open {
		if pak, err := cl.netPak.Pack(bytes); pak != nil {
			cl.pbPak <- pak
		} else if err != nil {
			fmt.Println(err)
		}
	} else if cl.st == st_close {
		return errors.New("连接状态异常")
	}
	return err
}

func (cl *client) Close() {
	cl.st = st_close
}
