package messenger

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type WSClientList map[*WSClient]bool

type WSClient struct {
	conn   *websocket.Conn
	app    *App
	egress chan []byte
}

func NewWSClient(conn *websocket.Conn, app *App) *WSClient {
	return &WSClient{
		conn:   conn,
		app:    app,
		egress: make(chan []byte, 50),
	}
}

func (c *WSClient) WaitForMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.app.removeWSClient(c)
	}()

	c.conn.SetPongHandler(c.pongHandler)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Errorf("error while setting read dealine: %v", err)
		return
	}

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Errorf("connection closed: %v", err)
				}
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Errorf("Error writing message to ws client: %v", err)
			}
			log.Debugf("sent message to ws client: %s", message)
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Errorf("error sending ping: %v", err)
				return
			}
		}
	}
}

func (c *WSClient) pongHandler(pongMsg string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
