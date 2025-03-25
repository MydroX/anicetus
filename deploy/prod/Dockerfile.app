FROM golang:1.24.1-alpine3.21 AS build

WORKDIR /app

COPY . .

RUN go build -o /bin/app ./cmd/main.go

FROM alpine:3.14.2
COPY --from=build /bin/app /bin/app
CMD [ "/bin/app" ]