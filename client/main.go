package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"../proto"
)

func main(){
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("failed to connect: %s", err)
	}

	client := proto.NewAddServiceClient(conn)

	message := proto.Request{
		A:2,
		B:4,
	}

	response, err := client.Add(context.Background(), &message)
	if err != nil{
		log.Fatalf("error when calling Add: %s", err)
	}
	//log.Print(response.Result)
	log.Printf("Response from server: %d", response.Result)
}