# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.3.3
MAINTAINER Benoit Hirbec <benoit.hirbec@gmail.com>

# Copy the local package files to the container's workspace.
COPY . /go/src

# Build the bitbot command inside the container.
RUN go install bitbot/cmd/...
