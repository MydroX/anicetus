FROM golang:1.24.1-alpine3.21

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.7

COPY . .

CMD ["air", "-c", "air.toml"]