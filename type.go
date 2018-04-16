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
	Data   []byte
	Silent bool
}

// RqdataSet - стркутура объекта set
type RqdataSet struct {
	Key string
	Obj ramstore.Obj
}

// RqdataGet - стркутура объекта get
type RqdataGet struct {
	Key string
}

// RqdataNotify - стркутура объекта нотификации
type RqdataNotify struct {
	Keys []string
	Data []byte
}

// Ansdata - Структера ответа
type Ansdata struct {
	Error string
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

type newSubscriber struct {
	Key  string
	Gw   *gob.Encoder
	Conn net.Conn
}
