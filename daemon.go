package ramnet

import (
	"log"
	"time"

	"github.com/fe0b6/config"
)

var (
	newSubscribe chan newSubscriber
	newNotify    chan RqdataNotify
)

func transmitDaemon(transmitChan chan Rqdata) {
	for d := range transmitChan {
		if d.Silent {
			continue
		}
		d.Silent = true

		for i := range clients {
			go clients[i].Send(d)
		}
	}
}

func syncDaemon() {
	// Синхронизация при зхапуске
	for _, addr := range config.GetStrArr("net", "route") {
		c := ClientConn{Addr: addr}
		c.sync()
	}
}

func notifyDaemon() {

	newNotify = make(chan RqdataNotify, 100)
	newSubscribe = make(chan newSubscriber, 10)

	conns := map[string][]newSubscriber{}

	for {
		select {
		case nc := <-newSubscribe:
			cs, ok := conns[nc.Key]
			if !ok {
				cs = []newSubscriber{}
			}

			cs = append(cs, nc)
			conns[nc.Key] = cs

		case n := <-newNotify:
			for _, k := range n.Keys {
				cs, ok := conns[k]
				if !ok {
					continue
				}

				ncs := []newSubscriber{}
				for _, c := range cs {
					err := c.Conn.SetWriteDeadline(time.Now().Add(ConnectTimeout))
					if err != nil {
						log.Println("[error]", err)
						continue
					}
					err = c.Gw.Encode(n.Data)
					if err != nil {
						log.Println("[error]", err)
						continue
					}

					ncs = append(ncs, c)
				}

				if len(ncs) != len(cs) {
					conns[k] = ncs
				}
			}
		}
	}

}
