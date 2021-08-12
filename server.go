package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//настройки websocket`а
var addr = flag.String("addr", "localhost:8080", "http service address")

//созданние сокета и установка его параметров
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var channel = make(chan dataIn)     //канал для передачи полученного сообщения
var allClient = []*websocket.Conn{} //слайс подключенных киентов

//входящие данные
type dataIn struct {
	data   []byte
	typ    int
	sender *websocket.Conn
}

type Params struct {
	Ids     string `json:"ids"`
	Message string `json:"message"`
}

//запрос
type requestMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Param   Params `json:"params"`
	Id      int    `json:"id"`
}

//ответ
type responseMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      int    `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//ответ ошибкой
type responseErrMsg struct {
	Jsonrpc string `json:"jsonrpc"`
	Jerror  Error  `json:"error"`
	Id      int    `json:"id"`
}

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
	go func(connect *websocket.Conn, channelIO chan dataIn) {
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
			data := dataIn{message, msgT, connect}
			channelIO <- data
			print(connect)
			fmt.Println(" put msg in ch")
		}

	}(connect, channel)
}

func sendMessage(myResponseMsg responseMsg, msgT int, cients []*websocket.Conn) error {
	msg, err := json.Marshal(myResponseMsg)
	if err != nil {
		fmt.Println("json Marshal err: ", err)
		return err
	}
	//проход по всем киентам
	for num := range cients {
		err = cients[num].WriteMessage(msgT, msg)
		if err != nil {
			log.Println("WriteMessage err:", err)
			return err
		}
		print(cients[num])
		fmt.Println(" Resend")
	}
	return err
}

func sendMessageAboutError(myResponseMsg responseErrMsg, msgT int, cient *websocket.Conn) error {
	msg, err := json.Marshal(myResponseMsg)
	if err != nil {
		fmt.Println("json Marshal err: ", err)
		return err
	}
	err = cient.WriteMessage(msgT, msg)
	if err != nil {
		log.Println("WriteMessage err:", err)
		return err
	}
	print(cient)
	fmt.Println(" Resend")

	return err
}

//механизм оповещений
func annunciator() {
	for {
		select {
		case newMsg := <-channel:
			var myRequestMsg requestMsg
			//пытаемся получить структуру
			if err := json.Unmarshal(newMsg.data, &myRequestMsg); err != nil {
				fmt.Println("json Unmarshal err: ", err)
				//Вызов процедуры с неправильной структурой
				fmt.Println("switch default")
				var myResponseErrMsg responseErrMsg
				myResponseErrMsg.Id = myRequestMsg.Id
				myResponseErrMsg.Jsonrpc = "2.0"
				myResponseErrMsg.Jerror.Code = -32600
				myResponseErrMsg.Jerror.Message = "Invalid JSON-RPC."
				if err := sendMessageAboutError(myResponseErrMsg, newMsg.typ, newMsg.sender); err != nil {
					fmt.Println("sendMessage err: ", err)
				}
			} else {
				fmt.Println("We got request for:", myRequestMsg.Method)
				var myResponseMsg responseMsg
				myResponseMsg.Id = myRequestMsg.Id
				//реакция на метод
				switch myRequestMsg.Method {
				case "sendMessage":
					fmt.Println("sendMessage")
					myResponseMsg.Jsonrpc = "2.0"
					myResponseMsg.Result = myRequestMsg.Param.Message
					if err := sendMessage(myResponseMsg, newMsg.typ, allClient); err != nil {
						fmt.Println("sendMessage err: ", err)
					}
				case "sendEcho":
					fmt.Println("sendEcho")
					myResponseMsg.Jsonrpc = "2.0"
					myResponseMsg.Result = myRequestMsg.Param.Message
					var echoclient = []*websocket.Conn{}
					echoclient = append(echoclient, newMsg.sender)
					if err := sendMessage(myResponseMsg, newMsg.typ, echoclient); err != nil {
						fmt.Println("sendMessage err: ", err)
					}
				default:
					fmt.Println("switch default")
					var myResponseErrMsg responseErrMsg
					myResponseErrMsg.Id = myRequestMsg.Id
					myResponseErrMsg.Jsonrpc = "2.0"
					myResponseErrMsg.Jerror.Code = -32601
					myResponseErrMsg.Jerror.Message = "Procedure not found."
					if err := sendMessageAboutError(myResponseErrMsg, newMsg.typ, newMsg.sender); err != nil {
						fmt.Println("sendMessage err: ", err)
					}
				}
			}
		}
	}
}

func main() {
	flag.Parse()     //анализ флагов
	go annunciator() //оповещение всех клиентов о новом сообщении
	fmt.Println("Run my websocket server")
	http.HandleFunc("/", wsEndpoint) //связываемся с обработчиком
	log.Fatal(http.ListenAndServe(*addr, nil))
}
