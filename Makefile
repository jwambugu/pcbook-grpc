.PHONY: gen-protos clean-protos

gen-protos:
	 protoc -I protos/ protos/*.proto --go_out=protos --go-grpc_out=protos

clean-protos:
	rm -f pb/*.go