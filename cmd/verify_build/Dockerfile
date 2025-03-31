# This Dockerfile builds an image with all the tools needed to
# reproduce firmware builds from manifest data for the armored
# witness.
# The default entrypoint runs a continuous monitor that attempts
# to build all firmware artifacts committed to by a log.
FROM golang:1.24-alpine AS builder

ARG GOFLAGS=""
ENV GOFLAGS=$GOFLAGS
ENV GO111MODULE=on

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o /build/bin/verifier ./cmd/verify_build

#
# Set up the final image
#
FROM ubuntu:22.04

RUN apt-get update
RUN apt-get -y install curl unzip wget xxd git

# Tamago requirements (versions of tamago will be installed at runtime)
RUN apt-get -y install binutils-arm-none-eabi build-essential make u-boot-tools musl-tools

ENV GOPATH /go
ENV PATH $GOPATH/bin:/testbase/protoc/bin:/usr/local/go/bin:$PATH

# Set up the proto compilation stuff
RUN cd /usr && \
    wget "https://github.com/google/protobuf/releases/download/v3.12.4/protoc-3.12.4-linux-x86_64.zip" && \
    unzip "protoc-3.12.4-linux-x86_64.zip"

COPY --from=builder /build/bin/verifier /bin/verifier

ENTRYPOINT ["/bin/verifier", "continuous", "--state_file=/tmp/state", "--cleanup=false"]
