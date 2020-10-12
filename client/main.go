package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"../proto"
	"os"
	"bufio"
	"fmt"
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


	fmt.Println("press a to add, m to multiply")

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()

	if err != nil {
	fmt.Println(err)
	}

	

	switch char {
	case 'a':
	fmt.Println("a Key Pressed")
	response, err := client.Add(context.Background(), &message)
	if err != nil{
		log.Fatalf("error when calling Add: %s", err)
	}
	//log.Print(response.Result)
	log.Printf("Response from server: %d", response.Result)
	break
	
	case 'm':
	fmt.Println("m Key Pressed")
	response, err := client.Multiply(context.Background(), &message)
	if err != nil{
		log.Fatalf("error when calling Add: %s", err)
	}
	//log.Print(response.Result)
	log.Printf("Response from server: %d", response.Result)
	break
	}

}