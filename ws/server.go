package ws

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

type WsServer struct {
	// 消息长度
	clientMessageBuffer int

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.
	publishLimiter *rate.Limiter

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	clientsMu sync.Mutex
	Clients   map[*client]struct{}

	MsgCh chan *Packet
}

var uidGen *UniqueID = &UniqueID{}

func NewWsServer() *WsServer {
	ws := &WsServer{
		clientMessageBuffer: 16,
		logf:                log.Printf,
		Clients:             make(map[*client]struct{}),
		publishLimiter:      rate.NewLimiter(rate.Every(time.Microsecond*100), 8),
	}
	// ws.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	ws.serveMux.HandleFunc("/ws", ws.clinetHandler)

	return ws
}

func (ws *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.serveMux.ServeHTTP(w, r)
}

// 客户端连接服务器
func (ws *WsServer) clinetHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		ws.logf("%v", err)
	}
	defer c.Close(websocket.StatusInternalError, "")

	err = ws.messageHandle(r.Context(), c)

	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		ws.logf("%v", err)
		return
	}
}

// 处理消息交互
func (ws *WsServer) messageHandle(ctx context.Context, c *websocket.Conn) error {
	cl := NewClient(uidGen.get(), ws.clientMessageBuffer, func() {
		c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
	})
	ws.addClient(cl)
	defer ws.deleteClient(cl)

	go ws.receiveMessage(ctx, c, cl)

	for {
		select {
		case err := <-cl.rcvCh:
			if err != nil {
				return err
			}
		case msg := <-cl.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg, websocket.MessageBinary)
			if err != nil {
				return err
			}
		case txt := <-cl.txts:
			err := writeTimeout(ctx, time.Second*5, c, []byte(txt), websocket.MessageText)
			if err != nil {
				return err
			}
		case pak := <-cl.pbPak:
			ws.recvPbPak(pak)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ws *WsServer) receiveMessage(ctx context.Context, c *websocket.Conn, cl *client) {
	for {
		err := cl.Read(ctx, c)
		if err != nil {
			// 判定连接是否关闭了，正常关闭，不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return
			} else if errors.Is(err, io.EOF) {
				return
			}
			cl.rcvCh <- err
			return
		}
	}
}

// 向所有客户端推送消息
func (ws *WsServer) ToA(msg []byte) {
	ws.clientsMu.Lock()
	defer ws.clientsMu.Unlock()

	ws.publishLimiter.Wait(context.Background()) // ???

	for c := range ws.Clients {
		select {
		case c.msgs <- msg:
		default:
			go c.closeSlow() // ???
		}
	}
}

func (ws *WsServer) ToC(session int64, msg []byte) {
	ws.clientsMu.Lock()
	defer ws.clientsMu.Unlock()

	for c := range ws.Clients {
		if c.session == session {
			c.msgs <- msg
			break
		}
	}
}

func (ws *WsServer) addClient(c *client) {
	ws.clientsMu.Lock()
	ws.Clients[c] = struct{}{}
	ws.clientsMu.Unlock()
}

func (ws *WsServer) deleteClient(c *client) {
	ws.clientsMu.Lock()
	delete(ws.Clients, c)
	ws.clientsMu.Unlock()
}

func (ws *WsServer) recvPbPak(pak *Packet) {
	if ws.MsgCh != nil {
		ws.MsgCh <- pak
	}
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte, tp websocket.MessageType) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, tp, msg)
}
