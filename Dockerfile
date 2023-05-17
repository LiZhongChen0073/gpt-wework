FROM golang:1.19-alpine

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

CMD ["./gpt-wework"]