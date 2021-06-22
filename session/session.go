package session

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

var (
	usernameHasBeenTaken = "username %s is already taken. please retry with a different name"
	retryMessage         = "failed to connect. please try again"
	welcome              = "Hello %s! \n Type your message in the chat"
	joined               = "%s has joined the chat!"
	chat                 = "%s: %s"
	left                 = "%s has left the chat!"
)
var Peers map[string]*websocket.Conn

func init() {
	Peers = map[string]*websocket.Conn{}
}

type Session struct {
	username string
	peer     *websocket.Conn
}

func NewSession(username string, peer *websocket.Conn) *Session {
	return &Session{username: username, peer: peer}
}
func (s *Session) Start() {
	usernameTaken, err := CheckUserNameExists(s.username)
	if usernameTaken {
		msg := fmt.Sprintf(usernameHasBeenTaken, s.username)
		s.peer.WriteMessage(websocket.TextMessage, []byte(msg))
		s.peer.Close()
		return
	}

	err = CreateUser(s.username)
	if err != nil {
		log.Println("failed to add user to list of active chat users", s.username)
		s.notifyPeer(retryMessage)
		s.peer.Close()
		return
	}
	
	Peers[s.username] = s.peer

	s.notifyPeer(fmt.Sprintf(welcome, s.username))
	SendToChannel(fmt.Sprintf(joined, s.username))

	go func() {
		log.Printf("%s joined to the chat\n", s.username)
		for {
			_, msg, err := s.peer.ReadMessage()
			if err != nil {
				_, ok := err.(*websocket.CloseError)
				if ok {
					log.Println("connection closed by user")
					s.disconnect()
				}
				return
			}
			SendToChannel(fmt.Sprintf(chat, s.username, string(msg)))
		}
	}()
}

func (s *Session) notifyPeer(msg string) {
	err := s.peer.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("failed to write message", err)
	}
}

func (s *Session) disconnect() {
	RemoveUser(s.username)

	SendToChannel(fmt.Sprintf(left, s.username))

	s.peer.Close()

	delete(Peers, s.username)
}
