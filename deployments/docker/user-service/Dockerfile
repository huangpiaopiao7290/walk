# 继承base dockerfile
FROM walk/base-go-app:latest

WORKDIR /app/user-service

COPY . .

RUN mkdir -p ${OUTPUT_DIR} && go build -o ${OUTPUT_DIR}/user-service

