############################
# Build
############################
FROM golang:1.18-bullseye AS build

# Ensure ca-certficates are up to date
RUN update-ca-certificates

# Set the current Working Directory inside the container
RUN mkdir /token-verifier
WORKDIR /token-verifier

# Use Go Modules
COPY go.mod .
COPY go.sum .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' -a \
       -o /go/bin/token-verifier

############################
# Image
############################
# using static nonroot image
# user:group is nonroot:nonroot, uid:gid = 65532:65532
FROM gcr.io/distroless/static@sha256:23aa732bba4c8618c0d97c26a72a32997363d591807b0d4c31b0bbc8a774bddf

EXPOSE 80/tcp

# Copy the Pre-built binary file from the previous stage
COPY --from=build /go/bin/token-verifier /run/token-verifier

ENTRYPOINT ["/run/token-verifier"]
