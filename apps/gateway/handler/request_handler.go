package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	etcd "go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/grpc/health/grpc_health_v1"

	gw_config "walk/apps/gateway/config"
	utils "walk/apps/gateway/utils"
)
type DynamicHandler struct {
	etcdClient 	*etcd.Client
	connCache 	sync.Map		// {key: serviceName, value: *grpc.ClientConn}
}

// @brief: 创建DynamicHandler实例
// @param: etcdClient *etcd.Client etcd客户端
// @return: *DynamicHandler DynamicHandler实例
func NewDynamicHandler(etcdClient *etcd.Client) *DynamicHandler {
	return &DynamicHandler{
		etcdClient: etcdClient,
	}
}

// @吧brief: 关闭所有缓存的 gRPC 连接
func (h *DynamicHandler) CloseConnections() {
	h.connCache.Range(func(key, value any) bool {
		conn := value.(*grpc.ClientConn)
		_ = conn.Close()
		return true
	})
}

// @brief: 获取grpc连接
// @param: serviceName 服务名称
// @return: *grpc.ClientConn gRPC连接
// @return: error 错误信息
func (h *DynamicHandler) getGRPCConn(serviceName string) (*grpc.ClientConn, error) {
	// 检查缓存
	if conn, ok := h.connCache.Load(serviceName); ok {
		if isHealthy(conn.(*grpc.ClientConn)) {
			return conn.(*grpc.ClientConn), nil
		}
		// 如果连接不健康，删除旧连接，准备重建
		h.connCache.Delete(serviceName)
	}

	// 从etcd查询服务地址
	resp, err := h.etcdClient.Get(context.TODO(), serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service from etcd: %v", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	addr := string(resp.Kvs[0].Value)

	// 建立gRPC连接
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %v", err)
	}
	// 确保新连接是健康的
	if !isHealthy(conn) {
		_ = conn.Close()
		return nil, fmt.Errorf("newly created connection is unhealthy")
	}

	// 缓存连接
	h.connCache.Store(serviceName, conn)
	return conn, nil
}

// isHealthy 检查 gRPC 连接是否健康
func isHealthy(conn *grpc.ClientConn) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil || resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return false
	}
	return true
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
// @param: ctx context.Context 上下文
// @param: conn *grpc.ClientConn gRPC连接
// @param: endpoint gw_config.EndpointItem gRPC服务端点
// @return: []byte 响应数据
// @return: error 错误信息
func (h *DynamicHandler) invokeGRPCMethod(ctx context.Context, 
							conn *grpc.ClientConn, endpoint gw_config.EndpointItem, reqBytes []byte) ([]byte, error) {
	// 方法描述
	fullMethod := fmt.Sprintf("/%s/%s", endpoint.GrpcService, endpoint.GrpcMethod)
	
    // 动态解析请求类型
	reqTypeName := fmt.Sprintf("%s.%sRequest",
		strings.ReplaceAll(endpoint.GrpcService, ".", ""),
		endpoint.GrpcMethod,
	)

	// 创建请求对象
	reqType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(reqTypeName))
	if err != nil {
		return nil, status.Errorf(codes.Unimplemented, "failed to find request type: %v", err)
	}

	req := reqType.New().Interface().(proto.Message)
	if err := proto.Unmarshal(reqBytes, req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal request: %v", err)
	}

	// 执行grpc调用
	var resp proto.Message
	if err := conn.Invoke(ctx, fullMethod, req, &resp); err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.Unavailable || st.Code() == codes.Internal {
			// 尝试重连
			newConn, reconnectErr := h.getGRPCConn(endpoint.GrpcService)
			if reconnectErr != nil {
				return nil, status.Errorf(codes.Unavailable, "reconnect failed: %v", reconnectErr)
			}
			err = newConn.Invoke(ctx, fullMethod, req, resp)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "retry after reconnect failed: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "failed to invoke gRPC method: %v", err)
		}
	}

	// 5.序列化响应
	return proto.Marshal(resp)
}

func (h *DynamicHandler) HandleGRPCRequest(endpoint *gw_config.EndpointItem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 服务发现
		serviceName := endpoint.GrpcService
		target, err := h.discoverService(serviceName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, utils.ErrorResponse(codes.Unavailable, err))
			return
		}

		// 建立gRPC连接
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		conn, err := h.getGRPCConn(target)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.ErrorResponse(codes.Internal, err))
			return
		}

		// 获取请求体
		req, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.ErrorResponse(codes.InvalidArgument, fmt.Errorf("failed to read request body: %v", err)))
			return
		}
		
		// 处理请求转换
		respBytes, err := h.invokeGRPCMethod(ctx, conn, *endpoint, req)
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
