# 继承base dockerfile
FROM walk/base-go-app:latest

WORKDIR /app/user-service

COPY . .

RUN go build -o user-service

