FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o chatapp ./cmd/chatapp

EXPOSE 8082

CMD ["./chatapp"]

