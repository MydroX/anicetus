FROM golang:1.24.4-alpine3.22

WORKDIR /app

RUN go install github.com/air-verse/air@v1.62.0

COPY . .

CMD ["air", "-c", "air.toml"]