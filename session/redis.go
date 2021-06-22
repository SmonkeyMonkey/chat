package session

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

type Config struct {
	Redis struct {
		Port string
		Password string
	}
}

var (
	client *redis.Client
	sub    *redis.PubSub
)

const channel = "chat"
const users = "user-chat"

func init() {
	var conf Config
	if _, err := toml.DecodeFile("./config.toml", &conf); err != nil {
		log.Fatal("error decoding toml file", err)
	}
	redisAddr := "localhost:" + conf.Redis.Port
	log.Println("trying to connection to redis...")
	client = redis.NewClient(&redis.Options{Addr: redisAddr,Password: conf.Redis.Password})

	if _, err := client.Ping().Result(); err != nil {
		log.Fatal("connection to redis is failed")
	}

	log.Println("sucessed connection to redis on ", redisAddr)
	
	startSubscriber()
}
func startSubscriber() {
	go func() {
		log.Println("start subsciber")
		sub = client.Subscribe(channel)
		messages := sub.Channel()
		for message := range messages {
			for _, peer := range Peers {
				peer.WriteMessage(websocket.TextMessage, []byte(message.Payload))
			}
		}
	}()
}

func SendToChannel(message string) {
	err := client.Publish(channel, message).Err()
	if err != nil {
		log.Println("Error publish to channel", err)
	}
}
func CheckUserNameExists(user string) (bool, error) {
	usernameTaken, err := client.SIsMember(users, user).Result()
	if err != nil {
		return false, err
	}
	return usernameTaken, nil
}
func CreateUser(user string) error {
	err := client.SAdd(users, user).Err()
	if err != nil {
		return err
	}
	return nil
}

func RemoveUser(user string) {
	err := client.SRem(users, user).Err()
	if err != nil {
		log.Println("error removing user: ", user)
		return
	}
	log.Printf("User %s is sucessed removing from redis", user)
}
func Clean() {
	for user, peer := range Peers {
		client.SRem(users, user)
		peer.Close()
	}
	log.Println("cleaned users sessions")
	err := sub.Unsubscribe(channel)
	if err != nil {
		log.Println("failed to unsubscribe redis channel subscription:", err)
	}
	err = sub.Close()
	if err != nil {
		log.Println("failed to close redis channel subscription:", err)
	}

	err = client.Close()
	if err != nil {
		log.Println("failed to close connection with redis: ", err)
		return
	}
}
