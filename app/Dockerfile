FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .. .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl bash
WORKDIR /root/

ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["/wait-for-it.sh", "postgres:5432", "--", "./main"]
