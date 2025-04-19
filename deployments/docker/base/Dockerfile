# base image
FROM archlinux:5.15.167

ENV DEBIAN_FRONTEND=noninteractive
ENV OUTPUT_DIR /app/bin

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm go protobuf

# go环境变量
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOPROXY https://goproxy.cn,direct

# 下载gRPC和protoc插件
RUN go get -u google.golang.org/grpc
RUN go get -u github.com/golang/protobuf/protoc-gen-go

# 创建工作目录
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# 拷贝源码
COPY . .

RUN go build -o app

# 构建镜像
# cd /home/pp/programs/program_go/timeTrack/walk/deployments/docker/base
# docker build -t your-dockerhub-username/base-go-app:latest .
# docker push walk/base-go-app:latest
