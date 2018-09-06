FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/github.com/maxbrunsfeld/counterfeiter
COPY . ./
RUN ./scripts/deps.sh
ARG VERSION=UNKNOWN
RUN go install -ldflags "-X main.AppVersion=${VERSION}" .

# Use golang as base image so go generate can be used directly
FROM golang:alpine
COPY --from=builder /go/bin/counterfeiter /usr/bin/counterfeiter
