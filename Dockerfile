# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o magicauth ./cmd/magicauth

FROM scratch

WORKDIR /root/

COPY --from=builder /app/magicauth .

EXPOSE 8080

CMD ["./magicauth"]
