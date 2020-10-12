package main

import (
	"container/list"
	"context"
	"log"
	"net"
	"time"

	"../proto"
	guuid "github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

type registroOrden struct {
	timestamp      time.Time
	id             guuid.UUID
	tipo           string
	nombreProducto string
	valor          int64
	origen         string
	destino        string
	seguimiento    guuid.UUID
}

type paquete struct {
	id           guuid.UUID
	tipo         string
	valor        int64
	origen       string
	destino      string
	intentos     int
	fechaEntrega time.Time
}

//crear las colas
var retail = list.New()
var prioritario = list.New()
var normal = list.New()

//lista de ordenes
var registroOrdenes = list.New()

func main() {
	//conexión
	listener, err := net.Listen("tcp", ":4040")
	if err != nil {
		log.Fatalf("failed to listen on port 4040: %v", err)
	}

	srv := grpc.NewServer()
	proto.RegisterAddServiceServer(srv, &server{})
	reflection.Register(srv)

	if e := srv.Serve(listener); e != nil {
		log.Fatalf("failed to Serve on port 4040: %v", e)
	}

}

func (s *server) Add(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	a, b := request.GetA(), request.GetB()
	result := a + b
	time.Sleep(1 * 60 * time.Second)
	return &proto.Response{Result: result}, nil
}

func (s *server) Multiply(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	a, b := request.GetA(), request.GetB()
	result := a * b

	return &proto.Response{Result: result}, nil
}

func (s *server) Order(ctx context.Context, request *proto.ClientRequest) (*proto.ResponseToClient, error) {
	id := guuid.New()
	seguimiento := guuid.New()
	orden := registroOrden{
		timestamp:      time.Now(),
		id:             id,
		tipo:           request.GetTipo(),
		nombreProducto: request.GetNombreProducto(),
		valor:          request.GetValor(),
		origen:         request.GetOrigen(),
		destino:        request.GetDestino(),
		seguimiento:    seguimiento,
	}
	registroOrdenes.PushBack(orden)
	paquete := paquete{
		id:       id,
		tipo:     request.GetTipo(),
		valor:    request.GetValor(),
		origen:   request.GetOrigen(),
		destino:  request.GetDestino(),
		intentos: 0,
	}
	switch request.GetTipo() {
	case "retail":
		retail.PushBack(paquete)
		break
	case "prioritario":
		prioritario.PushBack(paquete)
		break
	case "normal":
		normal.PushBack(paquete)
		break
	}

	return &proto.ResponseToClient{Seguimiento: seguimiento.String()}, nil
}
