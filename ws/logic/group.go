package logic

import (
	"context"
	"sync"
	"time"

	"github.com/BoelterC/utils/ws/msg"

	"golang.org/x/time/rate"
)

type Group struct {
	Gid  uint32
	Name string

	users map[int64]struct{}
	RcvCh chan *msg.Packet
}

func NewGroup(gid uint32) *Group {
	return &Group{
		Gid:   gid,
		users: map[int64]struct{}{},
		RcvCh: make(chan *msg.Packet, 32),
	}
}

func (g *Group) rmvUser(sid int64) {
	delete(g.users, sid)
}

type GroupManager struct {
	publishLimiter *rate.Limiter

	groups  map[uint32]*Group
	users   map[int64]*User
	usersMu sync.Mutex
	msgCh   chan *msg.Packet

	RcvCh chan *msg.Packet
}

var groupMgr *GroupManager

func NewGroupMgr() *GroupManager {
	if groupMgr == nil {
		groupMgr = &GroupManager{
			publishLimiter: rate.NewLimiter(rate.Every(time.Microsecond*100), 8),

			groups: map[uint32]*Group{},
			users:  make(map[int64]*User),
			msgCh:  make(chan *msg.Packet, 32),
			RcvCh:  make(chan *msg.Packet, 32),
		}
	}
	return groupMgr
}

func (g *GroupManager) Start() {
	for pak := range g.msgCh {
		if gp := g.groups[pak.Id]; gp != nil {
			gp.RcvCh <- pak
		} else {
			g.RcvCh <- pak
		}
	}
}

func (g *GroupManager) UserEnter(u *User) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	g.users[u.Session] = u
}

func (g *GroupManager) UserLeave(u *User) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	for _, gp := range g.groups {
		gp.rmvUser(u.Session)
	}
	delete(g.users, u.Session)
}

func (g *GroupManager) ToA(mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	g.publishLimiter.Wait(context.Background())

	for _, u := range g.users {
		select {
		case u.MsgCh <- &Msg{Pak: true, Mid: mid, Data: data}:
		default:
			go u.closeSlow()
		}
	}
}

func (g *GroupManager) ToG(gid uint32, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	if gp := g.groups[gid]; gp != nil {
	}
}

func (g *GroupManager) ToO(gid uint32, sid int64, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	if gp := g.groups[gid]; gp != nil {
	}
}

func (g *GroupManager) ToC(sid int64, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	if u := g.users[sid]; u != nil {
		u.MsgCh <- &Msg{Pak: true, Mid: mid, Data: data}
	}
}
