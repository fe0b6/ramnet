package ramnet

import (
	"encoding/gob"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fe0b6/ramstore"
)

func (c *clientConn) connet() (err error) {
	c.Lock()
	defer c.Unlock()
	if c.Connected {
		return
	}

	c.Conn, err = net.Dial("tcp", c.Addr)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") {
			log.Println("[error]", err)
		}
		return
	}
	c.Connected = true

	c.Gr = gob.NewDecoder(c.Conn)
	c.Gw = gob.NewEncoder(c.Conn)

	return
}

func (c *clientConn) reconnet() {
	for {
		err := c.connet()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Minute)
	}
}

func (c *clientConn) sync() {
	err := c.send(rqdata{Action: "sync"})
	if err != nil {
		log.Println("[error]", err)
		return
	}

	for {
		var ans ansdata
		c.Gr.Decode(&ans)
		if ans.EOF {
			break
		}

		ramstore.Set(ans.Key, ans.Obj)
	}
}

func (c *clientConn) send(d rqdata) (err error) {
	if !c.Connected {
		c.reconnet()
	}
	c.Lock()
	defer c.Unlock()

	err = c.Gw.Encode(d)
	if err != nil {
		log.Println("[error]", err)
		return
	}

	return
}

func transmit(d rqdata) {

	for i := range clients {
		clients[i].send(d)
	}

}
