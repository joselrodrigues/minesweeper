.PHONY: proto clean

# Variables
PROTO_DIR := proto
PROTO_FILE := minesweeper.proto
OUT_DIR := .

proto-go:
	protoc --go_out=$(OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/$(PROTO_FILE)

proto-py:
	python -m grpc_tools.protoc \
		-I$(PROTO_DIR) \
		--python_out=$(OUT_DIR) \
		--grpc_python_out=$(OUT_DIR) \
		$(PROTO_DIR)/$(PROTO_FILE)

clean:
	rm -f $(OUT_DIR)/*.pb.go
	rm -f $(OUT_DIR)/*_pb2.py
	rm -f $(OUT_DIR)/*_pb2_grpc.py

