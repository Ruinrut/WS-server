# Тестовый WS-сервер

 Вебсокет-сервер, который рассылает полученные сообщения всем подключенным клиентам. В качестве клиента использую [WebSocket Clien](https://marketplace.visualstudio.com/items?itemName=mohamed-nouri.websocket-client).

 Общение происходит через протокол jrpc. Реализованы методы:

- sendMessage - рассылка всем текстового сообщения;

- sendEcho - отпрравка сообщения самому себе.

### Сборка и запуск проекта
  Сборка Docker контейнера выполняется команой:
  
    docker build -t server .
 
 Запуск производиться командой:
 
    docker run --publish 6060:8080 -it server --addr 0.0.0.0:8080

Адрес для подключения клиента [ws://localhost:6060/](ws://localhost:6060/)