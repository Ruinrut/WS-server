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
	connect, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade err:", err)
		return
	}
	defer connect.Close() //закрыть соединение по завершению функции
	for {
		//получаем сообщение от клиента
		msgType, message, err := connect.ReadMessage()
		if err != nil {
			log.Println("ReadMessage err:", err)
			return
		}
		log.Printf("message: %s", message)
		//отправляем сообщение обратно
		err = connect.WriteMessage(msgType, message)
		if err != nil {
			log.Println("WriteMessage err:", err)
			return
		}
	}
}

func main() {
	fmt.Println("Run my websocket server")
	flag.Parse()                     //анализ флагов
	http.HandleFunc("/", wsEndpoint) //связываемся с обработчиком
	log.Fatal(http.ListenAndServe(*addr, nil))
}
