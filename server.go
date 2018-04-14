package ramnet

import (
	"encoding/gob"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fe0b6/config"
	"github.com/fe0b6/ramstore"
)

func runServer() (ln net.Listener) {
	var err error
	ln, err = net.Listen("tcp", config.GetStr("net", "host"))
	if err != nil {
		log.Fatalln("[error]", err)
		return
	}

	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				log.Println("[error]", err)
				continue
			}

			go handleServerConnection(conn)
		}
	}(ln)

	return
}

func handleServerConnection(conn net.Conn) {
	defer conn.Close()

	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	for {
		var d Rqdata
		err := gr.Decode(&d)
		if err != nil {
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "connection reset by peer") {
				log.Println("[error]", err)
			}
			break
		}

		var ans Ansdata
		switch d.Action {
		case "set":
			ans.Error = ramstore.Set(d.Key, d.Obj)
			if ans.Error == "" {
				go transmit(d)
			}
			log.Println("set", d.Key)

		case "get":
			ans.Obj, ans.Error = ramstore.Get(d.Key)
			log.Println("get", d.Key, ans.Error, string(ans.Obj.Data))

		case "del":
			if !d.Obj.Deleted {
				d.Obj = ramstore.Obj{
					Deleted: true,
					Time:    time.Now().UnixNano(),
				}
			}
			ans.Error = ramstore.Set(d.Key, d.Obj)
			if ans.Error == "" {
				go transmit(d)
			}

			log.Println("del", d.Key)

		case "sync":
			h := map[string]ramstore.Obj{}

			// Перебираем все элементы
			ramstore.Foreach(func(k string, v ramstore.Obj) {
				// Если объект nil то закончили обработку
				if k == "" {
					for n, d := range h {
						err = gw.Encode(Ansdata{Key: n, Obj: d})
						if err != nil {
							log.Println("[error]", err)
							continue
						}
					}

					h = map[string]ramstore.Obj{}
				}

				h[k] = v
			})

			ans.EOF = true

		default:
			log.Println("bad action", d)
			continue
		}

		log.Println("send ans", ans)

		err = gw.Encode(ans)
		if err != nil {
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "connection reset by peer") {
				log.Println("[error]", err)
			}
			continue
		}
	}

}
