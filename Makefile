gen-protos:
	 protoc -I protos/ protos/*.proto --go_out=protos --go-grpc_out=protos

clean-protos:
	rm -f pb/*.go

test:
	go test -cover -race ./...

run-server:
	go run cmd/server/main.go -port 8080

run-client:
	go run cmd/client/main.go -server-address 0.0.0.0:8080

gen-cert:
	cd certs; ./gen.sh; cd ..

.PHONY: gen-protos clean-protos test run-client run-server