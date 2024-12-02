FROM golang:1.23.3-alpine3.20

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.1

COPY . .

CMD ["air", "-c", "air.toml"]