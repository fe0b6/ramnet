package ramnet

import (
	"encoding/gob"
	"net"
	"sync"

	"github.com/fe0b6/ramstore"
)

type rqdata struct {
	Action string
	Key    string
	Obj    ramstore.Obj
	Sync   map[string]int64
}

type ansdata struct {
	Error error
	Key   string
	Obj   ramstore.Obj
	EOF   bool
}

type clientConn struct {
	Addr      string
	Connected bool
	Conn      net.Conn
	Gr        *gob.Decoder
	Gw        *gob.Encoder
	sync.Mutex
}
