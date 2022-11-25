package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BoelterC/utils/ws/msg"
	"github.com/BoelterC/utils/ws/util"

	"nhooyr.io/websocket"
)

const (
	st_close = iota
	st_init
	st_open
)

var u64 *util.UID64 = &util.UID64{}

type Msg struct {
	Pak  bool
	Gid  uint
	Mid  int
	Data []byte
}

type User struct {
	Session   int64        `json:"session"`
	Nickname  string       `json:"nickname"`
	EnterAt   time.Time    `json:"enter_at"`
	Addr      string       `json:"addr"`
	MsgCh     chan *Msg    `json:"-"`
	st        int          `json:"-"`
	handPak   *msg.HandPak `json:"-"`
	netPak    *msg.MsgPak  `json:"-"`
	closeSlow func()

	conn *websocket.Conn
}

func NewUser(conn *websocket.Conn, nickname, addr string, closeFunc func()) *User {
	return &User{
		Session:  u64.Get64(),
		Nickname: nickname,
		Addr:     addr,
		EnterAt:  time.Now(),
		MsgCh:    make(chan *Msg, 32),

		st:        st_init,
		handPak:   &msg.HandPak{},
		netPak:    &msg.MsgPak{},
		closeSlow: closeFunc,

		conn: conn,
	}
}

func (u *User) SendMsg(ctx context.Context) {
	for b := range u.MsgCh {
		if b.Pak {
			if data, err := msg.MakePacket(u.Session, b.Gid, b.Mid, b.Data); err == nil {
				u.conn.Write(ctx, websocket.MessageBinary, data)
			}
		} else {
			u.conn.Write(ctx, websocket.MessageText, b.Data)
		}
	}
}

func (u *User) Close(code websocket.StatusCode, reason string) {
	close(u.MsgCh)
	u.conn.Close(code, reason)
}

func recoverFromRecv(ctx context.Context) {
	if r := recover(); r != nil {
		fmt.Println(fmt.Printf("recoverFromRecv: %v", r))
		ctx.Done()
	}
}

func (u *User) RcvMsg(ctx context.Context) error {
	defer recoverFromRecv(ctx)

	for {
		_, bytes, err := u.conn.Read(ctx)
		if err != nil {
			return err
		}

		if u.st == st_init {
			result := u.handPak.RecvBuff(string(bytes))
			if result == msg.HEAD_ILLEGAL {
				u.st = st_close
				return errors.New("握手失败！非法的消息格式")
			} else if result == msg.HEAD_COMPLETE {
				u.st = st_open
				bytes, _ := json.Marshal(map[string]interface{}{"msg": "hand shake success!", "session": u.Session})
				u.MsgCh <- &Msg{Pak: false, Data: []byte(fmt.Sprintf("%s%s%s", msg.MSG_HEAD, bytes, msg.MSG_TAIL))}
			}
		} else if u.st == st_open {
			if pak, err := u.netPak.Pack(bytes); pak != nil {
				groupMgr.msgCh <- pak
			} else if err != nil {
				fmt.Println(err)
			}
		} else if u.st == st_close {
			return errors.New("连接状态异常")
		}
	}
}
