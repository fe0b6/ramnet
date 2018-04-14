package ramnet

import (
	"encoding/gob"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fe0b6/config"
	"github.com/fe0b6/ramstore"
	"github.com/fe0b6/tools"
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
			var obj RqdataSet
			tools.FromGob(&obj, d.Data)

			ans.Error = ramstore.Set(obj.Key, obj.Obj)
			if ans.Error == "" {
				go transmit(d)
			}

		case "multi_set":
			var objs []RqdataSet
			tools.FromGob(&objs, d.Data)

			for i := range objs {
				ans.Error = ramstore.Set(objs[i].Key, objs[i].Obj)
				if ans.Error != "" {
					break
				}
			}
			if ans.Error == "" {
				go transmit(d)
			}

		case "get":
			var obj RqdataGet
			tools.FromGob(&obj, d.Data)

			ans.Obj, ans.Error = ramstore.Get(obj.Key)
		case "multi_get":
			var objs []RqdataGet
			tools.FromGob(&objs, d.Data)

			for i := range objs {
				ans.Obj, ans.Error = ramstore.Get(objs[i].Key)
				err = gw.Encode(Ansdata{Key: objs[i].Key, Obj: ans.Obj})
				if err != nil {
					log.Println("[error]", err)
					break
				}
			}
			ans.EOF = true

		case "search":
			var obj RqdataGet
			tools.FromGob(&obj, d.Data)

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

				if k != "" && strings.Contains(k, obj.Key) {
					h[k] = v
				}
			})

			ans.EOF = true

		case "del":
			var obj RqdataSet
			tools.FromGob(&obj, d.Data)

			if !obj.Obj.Deleted {
				obj.Obj = ramstore.Obj{
					Deleted: true,
					Time:    time.Now().UnixNano(),
				}
			}

			ans.Error = ramstore.Set(obj.Key, obj.Obj)
			if ans.Error == "" {
				go transmit(d)
			}

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

		err = gw.Encode(ans)
		if err != nil {
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "connection reset by peer") {
				log.Println("[error]", err)
			}
			continue
		}
	}

}
