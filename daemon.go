package ramnet

import "github.com/fe0b6/config"

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
