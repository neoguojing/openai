 
# Use an official Golang runtime as a parent image
FROM golang:latest

# Install make command
RUN apt-get update && apt-get install -y make

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Build the Go app
RUN make build

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
WORKDIR /app/server
CMD ["./server"]


