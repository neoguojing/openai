
 
# Get the current directory
CUR_DIR := $(shell pwd)



.PHONY: build clean doc image

build:
	go build -o $(CUR_DIR)/chatbot/ $(CUR_DIR)/chatbot/
	go build -o $(CUR_DIR)/server/ $(CUR_DIR)/server/

clean:
	rm -f  $(CUR_DIR)/chatbot/chatbot
	rm -f  $(CUR_DIR)/server/server
	
doc:
	cd $(CUR_DIR)/server && swag init --parseDependency

 
image:
	# Build Docker image
	docker build -t guojingneo/chat-server:latest .
# Run Docker container
run:
	docker run -d -p 8080:8080 guojingneo/chat-server:latest







	


	



