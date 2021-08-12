package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWsEndpoint(t *testing.T) {
	//набор тестовых данных
	testClass := []struct {
		nane string
		req  string
		want string
	}{
		{
			nane: "sendMessage",
			req: "{\"jsonrpc\": \"2.0\",\"method\": \"sendMessage\",\"params\":" +
				" {\"ids\": \"*\",\"message\": \"Всем привет\"},\"id\": 3}",

			want: "{\"jsonrpc\":\"2.0\",\"result\":\"Всем привет\",\"id\":3}" +
				"{\"jsonrpc\":\"2.0\",\"result\":\"Всем привет\",\"id\":3}",
		},
		{
			nane: "sendEcho",
			req: "{\"jsonrpc\": \"2.0\",\"method\": \"sendEcho\"," +
				"\"params\": {\"message\": \"Сообщение себе\"},\"id\": 4}",

			want: "{\"jsonrpc\":\"2.0\",\"result\":\"Сообщение себе\",\"id\":4}",
		},
		{
			nane: "bad Procedure",
			req: "{\"jsonrpc\": \"2.0\",\"method\": \"update\"," +
				"\"params\": {\"message\": \"Сообщение себе\"},\"id\": 4}",

			want: "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32601," +
				"\"message\":\"Procedure not found.\"},\"id\":4}",
		},
		{
			nane: "bad JSON-RPC",
			req:  "Hello, WebServer!",

			want: "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32600," +
				"\"message\":\"Invalid JSON-RPC.\"},\"id\":0}",
		},
	}
	//подключаем наш поинт
	handler := http.HandlerFunc(wsEndpoint)
	//запускаем механизм оповещений
	go annunciator()
	for _, tc := range testClass {
		t.Run(tc.nane, func(t *testing.T) {
			//запускаем серрвер и настраеваем вебсокет
			server := httptest.NewServer(handler)
			defer server.Close()
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/"
			//поключаем клиента
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
			}

			res2 := new(string) //сообщение второго клиента
			
			go func(wsURL string, res *string) {
				//поключаем второго клиента
				ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err != nil {fmt.Println("could not open a ws connection on ", wsURL, err)}

				//получение ответа
				_, r, err := ws2.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						fmt.Printf("error: %v", err)
					}
				}

				*res = string(r)
			}(wsURL, res2)
			
			time.Sleep(1 * time.Millisecond)
			
			//отравка тестовых данных
			if err := ws.WriteMessage(websocket.TextMessage, []byte(tc.req)); err != nil {
				t.Fatalf("could not send message over ws connection %v", err)
			}
			//получение ответа
			_, res, err := ws.ReadMessage()
			if err != nil {
				t.Fatalf("%v", err)
			}
			time.Sleep(1 * time.Millisecond)

			//прповерка результата
			assert.Equal(t, tc.want, string(res)+*res2)
		})
	}
}
