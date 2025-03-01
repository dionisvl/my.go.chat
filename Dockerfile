FROM golang:1.23-alpine3.20

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY *.go ./
RUN go build -o main .

COPY ./.env* ./

EXPOSE 8080

CMD ["./main"]
