
 
# Get the current directory
CUR_DIR := $(shell pwd)



.PHONY: build clean

build:
	go build -o $(CUR_DIR)/chatbot/ $(CUR_DIR)/chatbot/
	go build -o $(CUR_DIR)/server/ $(CUR_DIR)/server/

clean:
	rm -f chatbot

# Generate godoc
godoc:
	godoc -http=:6060



