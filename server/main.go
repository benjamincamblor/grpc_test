package main

import(
	"log"
	"context"
	"net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"../proto"
	"time"
)

type server struct{}

func main(){
	listener, err := net.Listen("tcp",":4040")
	if err != nil{
		log.Fatalf("failed to listen on port 4040: %v", err)
	}

	srv := grpc.NewServer()
	proto.RegisterAddServiceServer(srv, &server{})
	reflection.Register(srv)

	if e := srv.Serve(listener); e!=nil{
		log.Fatalf("failed to Serve on port 4040: %v", e)
	}
}

func (s *server) Add(ctx context.Context, request *proto.Request) (*proto.Response, error){
	a,b := request.GetA(), request.GetB()
	result:= a+b
 	time.Sleep(1*60*time.Second)
	return &proto.Response{Result: result},nil
}

func (s *server) Multiply(ctx context.Context, request *proto.Request) (*proto.Response, error){
	a,b := request.GetA(), request.GetB()
	result:= a*b

	return &proto.Response{Result: result},nil
}