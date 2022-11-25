package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/BoelterC/utils/ws/logic"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

type Server struct {
	GroupMgr logic.GroupManager
	// publishLimiter controls the rate limit applied to the publish endpoint.
	// Defaults to one publish every 100ms with a burst of 8.
	publishLimiter *rate.Limiter
	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})
	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux
}

func NewWebsocket(logger func(f string, v ...interface{})) *Server {
	ws := &Server{
		GroupMgr:       *logic.NewGroupMgr(),
		logf:           logger,
		publishLimiter: rate.NewLimiter(rate.Every(time.Microsecond*100), 8),
	}
	// ws.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	ws.serveMux.HandleFunc("/ws", ws.userLinkHandler)
	return ws
}

// 客户端连接服务器
func (ws *Server) userLinkHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		ws.logf("%v", err)
	}

	user := logic.NewUser(conn, "", req.RemoteAddr, func() {
		conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
	})
	go user.SendMsg(req.Context())

	ws.GroupMgr.UserEnter(user)
	ws.logf("用户(%d) 连接服务器", user.Session)

	err = user.RcvMsg(req.Context())

	ws.GroupMgr.UserLeave(user)
	ws.logf("用户(%d) 与服务器断开", user.Session)

	if errors.Is(err, context.Canceled) {
		conn.Close(websocket.StatusNormalClosure, "")
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		conn.Close(websocket.StatusNormalClosure, "")
		return
	}

	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		ws.logf("Read from client error:", err.Error())
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}

func (ws *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.serveMux.ServeHTTP(w, r)
}
