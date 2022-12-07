package logic

import (
	"context"
	"sync"
	"time"

	"github.com/BoelterC/utils/ws/msg"
	"github.com/BoelterC/utils/ws/util"

	"golang.org/x/time/rate"
)

var uid32 *util.UID32 = &util.UID32{}

type Group interface {
	GetType() int
	GetId() uint
	SetId(gid uint)
	JoinGroup(sid int64, userinfo interface{}) bool
	LeaveGroup(sid int64) bool
	GetUsers() []int64
	GetRcvCh() chan *msg.Packet
	Dispose()
}

type GroupManager struct {
	publishLimiter *rate.Limiter

	groups  map[uint]Group
	gpsMu   sync.Mutex
	users   map[int64]*User
	usersMu sync.Mutex
	msgCh   chan *msg.Packet
}

var groupMgr *GroupManager

func NewGroupMgr() *GroupManager {
	if groupMgr == nil {
		groupMgr = &GroupManager{
			publishLimiter: rate.NewLimiter(rate.Every(time.Microsecond*100), 8),

			groups: map[uint]Group{},
			users:  make(map[int64]*User),
			msgCh:  make(chan *msg.Packet, 32),
		}
	}
	return groupMgr
}

func (g *GroupManager) Start(msgFunc func(*msg.Packet)) {
	for pak := range g.msgCh {
		if gp := g.groups[uint(pak.Id)]; gp != nil {
			gp.GetRcvCh() <- pak
		} else {
			if msgFunc != nil {
				msgFunc(pak)
			}
		}
	}
}

func (g *GroupManager) AddGroup(gp Group) uint {
	g.gpsMu.Lock()
	defer g.gpsMu.Unlock()

	gid := uint(uid32.Get32())
	gp.SetId(gid)
	g.groups[gid] = gp
	return gid
}

func (g *GroupManager) FindGroup(gid uint) Group {
	return g.groups[gid]
}

func (g *GroupManager) CloseGroup(gid uint) {
	g.gpsMu.Lock()
	defer g.gpsMu.Unlock()

	if gp := g.groups[gid]; gp != nil {
		gp.Dispose()
	}
	delete(g.groups, gid)
}

func (g *GroupManager) LeaveGroup(gid uint, sid int64) {
	g.gpsMu.Lock()
	defer g.gpsMu.Unlock()

	if gp := g.groups[gid]; gp != nil {
		gp.LeaveGroup(sid)
		if len(gp.GetUsers()) == 0 {
			g.CloseGroup(gid)
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
		g.LeaveGroup(gp.GetId(), u.Session)
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

func (g *GroupManager) ToG(gid uint, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	g.publishLimiter.Wait(context.Background())

	if gp := g.groups[gid]; gp != nil {
		for _, sid := range gp.GetUsers() {
			select {
			case g.users[sid].MsgCh <- &Msg{Pak: true, Gid: gid, Mid: mid, Data: data}:
			default:
				go g.users[sid].closeSlow()
			}
		}
	}
}

func (g *GroupManager) ToO(gid uint, sid int64, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	g.publishLimiter.Wait(context.Background())

	if gp := g.groups[gid]; gp != nil {
		for _, uid := range gp.GetUsers() {
			if uid == sid {
				continue
			}
			select {
			case g.users[sid].MsgCh <- &Msg{Pak: true, Gid: gid, Mid: mid, Data: data}:
			default:
				go g.users[sid].closeSlow()
			}
		}
	}
}

func (g *GroupManager) ToC(sid int64, mid int, data []byte) {
	g.usersMu.Lock()
	defer g.usersMu.Unlock()

	if u := g.users[sid]; u != nil {
		u.MsgCh <- &Msg{Pak: true, Mid: mid, Data: data}
	}
}
