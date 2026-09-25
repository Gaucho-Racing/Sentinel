FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
