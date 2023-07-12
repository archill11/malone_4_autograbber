FROM golang:1.20-alpine
WORKDIR /app
COPY ./ ./

# build go app
RUN go mod download
RUN go build -o myapp ./cmd/main.go

CMD ["./myapp"]