# map service image
FROM archlinux:5.15.167

ENV DEBIAN_FRONTEND=noninteractive

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm go protobuf

# go环境变量
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOPROXY https://goproxy.cn,direct

# 创建工作目录
WORKDIR /map-service

# 下载gRPC和protoc插件
RUN go get -u google.golang.org/grpc
RUN go get -u github.com/golang/protobuf/protoc-gen-go

# 拷贝proto文件
COPY proto/ /map-service/proto/
# 生成go文件
RUN protoc --go_out=plugins=grpc:. proto/*.proto

# 拷贝源码
COPY . /map-service

# 编译
RUN go build -o map-service

# 运行
CMD ["./map-service"]