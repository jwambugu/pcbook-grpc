gen-protos:
	 protoc -I protos/ protos/*.proto --go_out=protos --go-grpc_out=protos --grpc-gateway_out=protos --openapiv2_out=swagger

clean-protos:
	rm -f pb/*.go

test:
	go test -cover -race ./...

run-grpc-server:
	go run cmd/server/main.go -port 8080 -enable-tls -server-type=grpc

run-rest-server:
	go run cmd/server/main.go -port 8081 -enable-tls -server-type=rest

run-client:
	go run cmd/client/main.go -server-address 0.0.0.0:8080 -enable-tls

gen-cert:
	cd certs; ./gen.sh; cd ..

.PHONY: gen-protos clean-protos test run-client run-grpc-server run-rest-server gen-cert