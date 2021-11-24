.PHONY: gen-protos clean-protos

gen-protos:
	 protoc -I proto/ proto/*.proto --go_out=pb --go-grpc_out=pb

clean-protos:
	rm -f pb/*.go