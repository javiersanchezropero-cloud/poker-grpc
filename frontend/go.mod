module poker-grpc/frontend

go 1.22

require (
	poker-grpc/backend v0.0.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

replace poker-grpc/backend => ../backend
