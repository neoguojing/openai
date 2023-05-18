
 
# Get the current directory
CUR_DIR := $(shell pwd)



.PHONY: build clean

build:
	go build -o $(CUR_DIR)/chatbot/ $(CUR_DIR)/chatbot/

clean:
	rm -f chatbot

