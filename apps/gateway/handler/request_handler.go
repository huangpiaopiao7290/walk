package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	etcd "go.etcd.io/etcd/clientv3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	gw_config "walk/apps/gateway/config"
	utils "walk/apps/gateway/utils"
)
type DynamicHandler struct {
	etcdClient *etcd.Client
}

func NewDynamicHandler(etcdClient *etcd.Client) *DynamicHandler {
	return &DynamicHandler{
		etcdClient: etcdClient,
	}
}

// @brief: 服务发现
// @param: serviceName 服务名称
// @return: string 服务地址
// @return: error 错误信息
func (h *DynamicHandler) discoverService(serviceName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcdClient.Get(
		ctx,
		fmt.Sprintf("/services/%s", serviceName),
		etcd.WithPrefix(),
		etcd.WithSerializable(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get service from etcd: %v", err)
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("service %s not found", serviceName)
	}
	return string(resp.Kvs[0].Value), nil

}

// @brief: 将请求转换为gRPC请求，并调用相应的gRPC方法
// @param: c gin.Context 上下文
// @param: conn *grpc.ClientConn gRPC连接
// @param: endpoint gw_config.EndpointItem gRPC服务端点
// @return: []byte 响应数据
// @return: error 错误信息
func (h *DynamicHandler) invokeGRPCMethod(c *gin.Context, conn *grpc.ClientConn, endpoint gw_config.EndpointItem) ([]byte, error) {
	// 1.准备方法描述
	fullMethod := fmt.Sprintf("/%s/%s", endpoint.GrpcService, endpoint.GrpcMethod)
	
	// 2.解析请求体
	reqBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		 return nil, status.Errorf(codes.InvalidArgument, "failed to read request body: %v", err)
	}

	// 3.动态创建请求对象
	reqType, err := protoregistry.GlobalTypes.FindMessageByName(
		protoreflect.FullName(fmt.Sprintf("%s.%sRequest",
			strings.ReplaceAll(endpoint.GrpcService, ".", ""),
			endpoint.GrpcMethod,
		)),)
	if err != nil {
		return nil, status.Errorf(codes.Unimplemented, "failed to find request type: %v", err)
	}

	req := reqType.New().Interface().(proto.Message)
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal request: %v", err)
	}

	// 4.执行grpc调用
	var resp proto.Message
	if err := conn.Invoke(c, fullMethod, req, &resp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to invoke gRPC method: %v", err)
	}

	// 5.序列化响应
	return proto.Marshal(resp)
}

func (h *DynamicHandler) HandleGRPCRequest(endpoint *gw_config.EndpointItem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1.服务发现
		serviceName := endpoint.GrpcService
		target, err := h.discoverService(serviceName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, utils.ErrorResponse(codes.Unavailable, err))
			return
		}

		// 2.建立gRPC连接
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(
			ctx, 
			target, 
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		defer conn.Close()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.ErrorResponse(codes.Internal, err))
			return
		}

		// 3.处理请求转换
		respBytes, err := h.invokeGRPCMethod(c, conn, *endpoint)
		if err != nil {
			h.handleGRPCError(c, err)
			return
		}

		// 4.返回响应
		c.Data(http.StatusOK, "application/json", respBytes)
	}

}

// @brief: 统一处理gRPC错误
// @param: c gin.Context 上下文
// @param: err error 错误信息
func (h *DynamicHandler) handleGRPCError(c *gin.Context, err error) {
	if st, ok := status.FromError(err); ok {
		code := utils.GrpcCodeToHTTP(st.Code())
		c.AbortWithStatusJSON(code, utils.ErrorResponse(st.Code(), st.Err()))
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, 
			utils.ErrorResponse(codes.Unknown, err))
	}
}
