package ramnet

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fe0b6/config"
	"github.com/fe0b6/ramstore"
	"github.com/fe0b6/tools"
)

func runServer(transmitChan chan Rqdata) (ln net.Listener) {
	var err error
	ln, err = net.Listen("tcp", config.GetStr("net", "host"))
	if err != nil {
		log.Fatalln("[error]", err)
		return
	}

	go func(ln net.Listener, transmitChan chan Rqdata) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("[error]", err)
				break
			}

			go handleServerConnection(conn, transmitChan)
		}
	}(ln, transmitChan)

	return
}

func handleServerConnection(conn net.Conn, transmitChan chan Rqdata) {
	defer conn.Close()

	//checkReconnect(conn.RemoteAddr().String())

	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	for {
		var d Rqdata
		err := gr.Decode(&d)
		if err != nil {
			if err != io.EOF {
				log.Println("[error]", err)
			}
			break
		}

		// Отвечаем что получили запрос
		err = gw.Encode(true)
		if err != nil {
			log.Println("[error]", err)
			break
		}

		var ans Ansdata
		switch d.Action {
		case "set":
			var obj RqdataSet
			tools.FromGob(&obj, d.Data)

			if debug {
				log.Println("set", obj.Key, obj.Obj.Time)
			}

			ans.Error = ramstore.Set(obj.Key, obj.Obj)
			if ans.Error == "" {
				transmitChan <- d
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
				transmitChan <- d
			}

		case "get":
			var obj RqdataGet
			tools.FromGob(&obj, d.Data)

			ans.Obj, ans.Error = ramstore.Get(obj.Key)
		case "multi_get":
			var objs []RqdataGet
			tools.FromGob(&objs, d.Data)

			for i := range objs {
				ans.Obj, _ = ramstore.Get(objs[i].Key)
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

				if strings.Contains(k, obj.Key) {
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

			if debug {
				log.Println("del", obj.Key)
			}

			ans.Error = ramstore.Set(obj.Key, obj.Obj)
			if ans.Error == "" {
				transmitChan <- d
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
		case "ping":
			ans.EOF = true

		case "subscribe":
			var keys []string
			tools.FromGob(&keys, d.Data)
			d.Silent = true

			if debug {
				log.Println("subscribe", keys)
			}

			for _, k := range keys {
				newSubscribe <- newSubscriber{
					Key:  k,
					Gw:   gw,
					Conn: conn,
				}
			}

		case "notify":
			transmitChan <- d

			if debug {
				log.Println("notify", "get")
			}

			var n RqdataNotify
			tools.FromGob(&n, d.Data)
			newNotify <- n

			if debug {
				log.Println("notify", n.Keys)
			}

		default:
			log.Println("bad action", d)
			continue
		}

		if !d.Silent {
			err = gw.Encode(ans)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
		}
	}

}
