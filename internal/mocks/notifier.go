package mocks

import (
	"context"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-proto/go/notifierpb"
)

type NotifierOk struct {
}

func (n NotifierOk) CheckUser(ctx context.Context, in *notifierpb.CheckUserRequest, opts ...client.CallOption) (*notifierpb.CheckUserResponse, error) {
	return &notifierpb.CheckUserResponse{
		Status: 200,
	}, nil
}

func NewNotifierOk() notifierpb.NotifierService {
	return &NotifierOk{}
}
