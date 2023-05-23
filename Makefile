
 
# Get the current directory
CUR_DIR := $(shell pwd)
# amd64 | arm64
BUILD_ARCH ?= amd64
# amd64 | arm64 | cuda10
BUILD_DEVICE ?= amd64
PROJECT_VERSION := v1.1.0
BUILD := `git rev-parse --short HEAD`


.PHONY: build clean doc image

build:
	go build -o $(CUR_DIR)/chatbot/ $(CUR_DIR)/chatbot/
	go build -o $(CUR_DIR)/server/ $(CUR_DIR)/server/
	go build -o $(CUR_DIR)/wechat/ $(CUR_DIR)/wechat/

clean:
	rm -f  $(CUR_DIR)/chatbot/chatbot
	rm -f  $(CUR_DIR)/server/server
	rm -f  $(CUR_DIR)/wechat/wechat
	
doc:
	cd $(CUR_DIR)/server && swag init --parseDependency

 
image:
	# Build Docker image
	docker build -t guojingneo/chat-server:$(PROJECT_VERSION)-$(BUILD)-$(BUILD_ARCH)-$(BUILD_DEVICE) .
# Run Docker container
run:
	docker run -d -p 8080:8080 guojingneo/chat-server:$(PROJECT_VERSION)-$(BUILD)-$(BUILD_ARCH)-$(BUILD_DEVICE)







	


	



