package ramnet

import (
	"encoding/gob"
	"net"
	"sync"

	"github.com/fe0b6/ramstore"
)

// Rqdata - Структура запроса
type Rqdata struct {
	Action string
	Key    string
	Obj    ramstore.Obj
	Sync   map[string]int64
}

// Ansdata - Структера ответа
type Ansdata struct {
	Error error
	Key   string
	Obj   ramstore.Obj
	EOF   bool
}

// ClientConn - объект клиента
type ClientConn struct {
	Addr      string
	Connected bool
	Conn      net.Conn
	Gr        *gob.Decoder
	Gw        *gob.Encoder
	sync.Mutex
}
