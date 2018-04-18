package ramnet

import (
	"log"
	"net"

	"github.com/fe0b6/config"
)

var (
	clients []*ClientConn
	debug   bool
)

// Run - запуск сервера
func Run() (exitChan chan bool) {
	debug = config.GetBool("debug")

	go notifyDaemon()

	// Запускаем демон распространения
	transmitChan := make(chan Rqdata, 100)
	go transmitDaemon(transmitChan)

	// Создаем клиенты
	clients = []*ClientConn{}
	for _, addr := range config.GetStrArr("net", "route") {
		c := ClientConn{Addr: addr}
		clients = append(clients, &c)
	}

	// Запускаем сервер
	ln := runServer(transmitChan)

	// Канал для оповещения о выходе
	exitChan = make(chan bool)
	go waitExit(exitChan, ln)

	// Запускаем синхронизацию
	go syncDaemon()

	return
}

// Ждем сигнал о выходе
func waitExit(exitChan chan bool, ln net.Listener) {
	_ = <-exitChan

	log.Println("[info]", "Завершаем работу netserv ramkv")

	err := ln.Close()
	if err != nil {
		log.Println("[error]", err)
		return
	}

	log.Println("[info]", "Работа netserv ramkv завершена корректно")
	exitChan <- true
}
