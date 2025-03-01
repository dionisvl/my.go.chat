FROM golang:1.23-alpine3.20

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

COPY ./.env* ./

EXPOSE 8080

CMD ["./server"]
