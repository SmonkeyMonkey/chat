Golang chat for real-time communication based on gorilla/websocket which realise WebSocket protocol,redis for saving messages and BurntSushi/toml used for parsing configuration file
## Clone 
```
git clone github.com/smonkeymonkey/chat
```
config.toml has default settings,edit them if you need specific settings

## Run 
```
make
```
or
```
go run *.go
```

## Usage
Go on http://localhost:8080, type your username in window and you can speaking :)

## Features
Nicknames in chat has be unique
Notification about new user join/left
Ctrl+C in console removes chat history Set from Redis
