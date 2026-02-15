FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/mysql-dump-cleaner ./main.go

FROM alpine:3.21
RUN adduser -D -h /app appuser
WORKDIR /app
COPY --from=build /bin/mysql-dump-cleaner /usr/local/bin/mysql-dump-cleaner
USER appuser
ENTRYPOINT ["mysql-dump-cleaner"]
