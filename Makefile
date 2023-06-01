
 
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
	cp $(CUR_DIR)/config/config.yaml.template $(CUR_DIR)/chatbot/config.yaml
	cp $(CUR_DIR)/role/role.yaml $(CUR_DIR)/chatbot/role.yaml
	go build -o $(CUR_DIR)/server/ $(CUR_DIR)/server/
	cp $(CUR_DIR)/config/config.yaml.template $(CUR_DIR)/server/config.yaml
	cp $(CUR_DIR)/role/role.yaml $(CUR_DIR)/server/role.yaml
	go build -o $(CUR_DIR)/wechat/ $(CUR_DIR)/wechat/
	cp $(CUR_DIR)/config/config.yaml.template $(CUR_DIR)/wechat/config.yaml
	cp $(CUR_DIR)/role/role.yaml $(CUR_DIR)/wechat/role.yaml
	cp $(CUR_DIR)/config/config.yaml.template $(CUR_DIR)/docker-compose/config.yaml

clean:
	rm -f  $(CUR_DIR)/chatbot/chatbot
	rm -f  $(CUR_DIR)/server/server
	rm -f  $(CUR_DIR)/wechat/wechat
	
doc:
	cd $(CUR_DIR)/server && swag init --parseDependency

 
image:
	# Build Docker image
	docker build -t guojingneo/chat-server:$(PROJECT_VERSION)-$(BUILD)-$(BUILD_ARCH)-$(BUILD_DEVICE) -f Dockerfile.server .
	docker build -t guojingneo/bot:$(PROJECT_VERSION)-$(BUILD)-$(BUILD_ARCH)-$(BUILD_DEVICE) -f Dockerfile.bot .
# Run Docker container
run:
	docker run -d -p 8080:8080 guojingneo/chat-server:$(PROJECT_VERSION)-$(BUILD)-$(BUILD_ARCH)-$(BUILD_DEVICE)







	


	



