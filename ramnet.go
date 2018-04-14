package ramnet

import (
	"log"
	"net"

	"github.com/fe0b6/config"
)

var (
	clients []*ClientConn
)

// Run - запуск сервера
func Run() (exitChan chan bool) {

	// Запускаем сервер
	ln := runServer()

	// Канал для оповещения о выходе
	exitChan = make(chan bool)
	go waitExit(exitChan, ln)

	// Синхронизация при зхапуске
	for _, addr := range config.GetStrArr("net", "route") {
		c := ClientConn{Addr: addr}
		go c.sync()
	}

	clients = []*ClientConn{}
	// Синхронизация при зхапуске
	for _, addr := range config.GetStrArr("net", "route") {
		c := ClientConn{Addr: addr}
		clients = append(clients, &c)
	}

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
