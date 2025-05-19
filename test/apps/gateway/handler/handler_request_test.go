package handler_request_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type MockEtcdClient struct {
	serviceAddr string
}

func (m *MockEtcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if key == "testService" {
		return &clientv3.GetResponse{
			Kvs: []*clientv3.KeyValue{
				{Value: []byte(m.serviceAddr)},
			},
		}, nil
	}
	return nil, errors.New("service not found")
}


func TestGetGRPCConn(t *testing.T) {

}