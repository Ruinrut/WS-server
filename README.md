# Тестовый WS-сервер

 Вебсокет-сервер, который рассылает полученные сообщения всем подключенным клиентам. В качестве клиента использую [WebSocket Clien](https://marketplace.visualstudio.com/items?itemName=mohamed-nouri.websocket-client).

 Общение происходит через протокол jrpc. Реализованы методы:

- sendMessage - рассылка всем текстового сообщения;

- sendEcho - отпрравка сообщения самому себе.
