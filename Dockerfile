# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.3.3
MAINTAINER Benoit Hirbec <benoit.hirbec@gmail.com>

# Copy the local package files to the container's workspace.
ADD src /go/src

# Build the bitbot command inside the container.
RUN go install bitbot

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/bitbot

# Document that the service listens on port 8080.
EXPOSE 8080
