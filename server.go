package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//настройки websocket`а
var addr = flag.String("addr", "localhost:8080", "http service address")

//создданние сокета и установка его параметров
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//обработчик для шаблона
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Upgrade")
		return
	}
}

func main() {
	fmt.Println("Run my websocket server")
	flag.Parse() //анализ флагов
	http.HandleFunc("/", wsEndpoint) //связываемся с обработчиком
	log.Fatal(http.ListenAndServe(*addr, nil))
}
