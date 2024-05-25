FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o magicauth ./cmd/magicauth

FROM scratch

WORKDIR /root/

COPY --from=builder /app/magicauth .

EXPOSE 8080

CMD ["./magicauth"]
