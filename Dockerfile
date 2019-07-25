FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
