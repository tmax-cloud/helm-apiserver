package release

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	gsocket "github.com/gorilla/websocket"
	"github.com/tmax-cloud/helm-apiserver/internal/utils"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"k8s.io/klog"
)

func init() {
	hub = newHub()
	go hub.run()
}

type Client struct {
	hub *Hub

	conn *gsocket.Conn

	send chan schemas.ReleaseResponse

	sh *ReleaseHandler

	ns string
}

type Hub struct {
	clients map[*Client]string

	broadcast chan schemas.ReleaseResponse

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan schemas.ReleaseResponse),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]string),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = client.ns
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case releaseList := <-h.broadcast:
			for client, ns := range h.clients {
				message := filter(releaseList, ns)
				if len(message.Release) != 0 {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)

					}
				}
			}
		}
	}
}

// namespace filtering
func filter(releaseList schemas.ReleaseResponse, ns string) schemas.ReleaseResponse {
	response := schemas.ReleaseResponse{}
	if ns == "" {
		return releaseList
	}

	for _, rel := range releaseList.Release {
		if rel.Namespace == ns {
			response.Release = append(response.Release, rel)
		}
	}
	return response
}

var hub *Hub

// serveWs handles websocket requests from the peer.
func (sh *ReleaseHandler) Websocket(w http.ResponseWriter, r *http.Request) {
	klog.Info("Start websocket connection")
	conn, err := utils.UpgradeWebsocket(w, r)
	if err != nil {
		klog.Errorln(err)
		return
	}

	vars := mux.Vars(r)
	namespace := vars["ns-name"]

	client := &Client{hub: hub, conn: conn, send: make(chan schemas.ReleaseResponse, 256), sh: sh, ns: namespace}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage() // message 필요한 상황을 위해 남겨둠
		if err != nil {
			if gsocket.IsUnexpectedCloseError(err, gsocket.CloseGoingAway, gsocket.CloseAbnormalClosure) {
				klog.Info(err)
			}
			break
		}

		releaseList := c.sh.GetReleasesForWS(c.ns) // ReleaseList 받아오기

		respMsg, err := json.Marshal(releaseList)

		c.conn.WriteMessage(gsocket.TextMessage, respMsg)
		if err != nil {
			klog.Error(err)
			if gsocket.IsUnexpectedCloseError(err, gsocket.CloseGoingAway, gsocket.CloseAbnormalClosure) {
				klog.Error(err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if !ok {
			// The hub closed the channel.
			c.conn.WriteMessage(gsocket.CloseMessage, []byte{})
			return
		}
		t, err := json.Marshal(message)
		if err != nil {
			klog.Info(err)
			return
		}
		c.conn.WriteMessage(gsocket.TextMessage, t)
	}
}
