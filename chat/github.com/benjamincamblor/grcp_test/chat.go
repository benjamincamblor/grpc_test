package chat

import (
	//"context"
	"log"
	"golang.org/x/net/context"
)

type Server struct{

}

func (s *Server) SayHello(ctx context.Context, message *Message) (*Message, error){
	log.Printf("recieved message body from client: %s", message.body)
	return &Message{Body: "Hello from the server"}, nil
}