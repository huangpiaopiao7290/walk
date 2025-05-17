package utils

import (
	"google.golang.org/grpc/codes"
	"net/http"

	gw_pb "walk/apps/gateway/pb"
)

// 统一错误响应格式
func ErrorResponse(code codes.Code, err error) *gw_pb.Response {
	return &gw_pb.Response{
		StatusCode: int32(GrpcCodeToHTTP(code)),
		Errors:     []string{err.Error()},
	}
}

// gRPC状态码转HTTP状态码
func GrpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Internal, codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}