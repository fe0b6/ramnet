package ramnet

import (
	"encoding/gob"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fe0b6/ramstore"
)

// ConnectTimeout - время ожидания ответа
const ConnectTimeout = 500 * time.Millisecond

// Connet - подключаемся к серверу
func (c *ClientConn) Connet() (err error) {
	c.Lock()
	defer c.Unlock()
	if c.Connected {
		return
	}

	log.Println("try connect to " + c.Addr)

	c.Conn, err = net.Dial("tcp", c.Addr)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") {
			log.Println("[error]", err)
		}
		return
	}
	c.Connected = true

	log.Println("connected to " + c.Addr)

	c.Gr = gob.NewDecoder(c.Conn)
	c.Gw = gob.NewEncoder(c.Conn)

	return
}

func (c *ClientConn) reconnet() {
	for {
		c.Connected = false
		err := c.Connet()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Minute)
	}
}

func (c *ClientConn) sync() {
	err := c.Send(Rqdata{Action: "sync"})
	if err != nil {
		log.Println("[error]", err)
		return
	}

	for {
		var ans Ansdata
		c.Gr.Decode(&ans)
		if ans.EOF {
			break
		}

		ramstore.Set(ans.Key, ans.Obj)
	}
}

// Send - шлем данные
func (c *ClientConn) Send(d Rqdata) (err error) {
	if !c.Connected {
		c.reconnet()
	}
	c.Lock()
	defer c.Unlock()

	// Устанавливаем таймаут на запись
	err = c.Conn.SetWriteDeadline(time.Now().Add(ConnectTimeout))
	if err != nil {
		log.Println("[error]", err)
		return
	}

	err = c.Gw.Encode(d)
	if err != nil {
		if strings.Contains(err.Error(), "broken pipe") {
			c.reconnet()
			return c.Send(d)
		}
		log.Println("[error]", err)
		return
	}

	return
}

func transmit(d Rqdata) {

	for i := range clients {
		clients[i].Send(d)
	}

}
