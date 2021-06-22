package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/websocket"
	"github.com/smonkeymonkey/chat/session"
)

var (
	addr     string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

type Config struct {
	Server struct {
		Host string
		Port string
	}
}

func init() {
	var conf Config

	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		log.Fatal(err)
	}
	addr = conf.Server.Host + ":" + conf.Server.Port

}

func main() {
	server := http.Server{Addr: addr}
	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/chat/", wsHandler)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("failed start server")
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	session.Clean()

	server.Shutdown(ctx)

	log.Println("chat app is closed")
}

func wsHandler(rw http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/chat/")

	peer, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Fatal("failed connect to websocket")
	}

	session := session.NewSession(username, peer)
	session.Start()
}
