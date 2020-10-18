install instructions:
go get google.golang.org/grpc   
go get github.com/google/uuid
protoc --proto_path=proto --go_out=plugins=grpc:proto service.proto
go get github.com/streadway/amqp