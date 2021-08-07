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
var channelIO = make(chan []byte)   //канал для передачи сообщения для оповещения
var channelTypIO = make(chan int)   //канал для передачи типа сообщения
var allClient = []*websocket.Conn{} //слайс подключенных киентов

//обработчик для шаблона
func wsEndpoint(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	//подключение
	connect, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade err:", err)
		return
	}
	//добавление нового клиента
	allClient = append(allClient, connect)
	go func(connect *websocket.Conn, channelIO chan []byte) {
		print(connect)
		fmt.Println(" cient connect")
		defer connect.Close() //закрыть соединение по завершению функции

		for {
			//получаем сообщение от клиента
			msgT, message, err := connect.ReadMessage()
			if err != nil {
				log.Println("ReadMessage err:", err)
				return
			}
			//отправляем сообщение на оповещение
			channelIO <- message
			//отправляем его тип
			channelTypIO <- msgT
			print(connect)
			fmt.Println(" put msg in ch")
		}

	}(connect, channelIO)

}

func main() {
	fmt.Println("Run my websocket server")
	flag.Parse() //анализ флагов

	//оповещение всех клиентов о новом сообщении
	go func() {
		for {
			select {
			case msg := <-channelIO:
				msgT := <-channelTypIO
				//прохо по всем киентам
				for cient := range allClient {
					err := allClient[cient].WriteMessage(msgT, msg)
					if err != nil {
						log.Println("WriteMessage err:", err)
					}
					print(allClient[cient])
					fmt.Println(" Resend")
				}
			}
		}
	}()

	http.HandleFunc("/", wsEndpoint) //связываемся с обработчиком
	log.Fatal(http.ListenAndServe(*addr, nil))
}
