.PHONY: proto clean

# Variables
PROTO_DIR := proto
PROTO_FILE := minesweeper.proto
GO_OUT_DIR := .

proto: clean
	protoc --go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/$(PROTO_FILE)

clean:
	rm -f $(GO_OUT_DIR)/*.pb.go
