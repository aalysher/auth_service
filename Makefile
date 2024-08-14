gen-rpc:
	protoc --go_out=proto --go-grpc_out=proto proto/auth.proto


