package oci

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

type Client interface {
	GetSubnetById(context.Context, string) (*core.Subnet, error)
}

type ClientImpl struct {
	vnClient core.VirtualNetworkClient
}

func NewClient(provider common.ConfigurationProvider) (Client, error) {
	net, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	return &ClientImpl{
		vnClient: net,
	}, nil
}

func (c *ClientImpl) GetSubnetById(ctx context.Context, subnetId string) (*core.Subnet, error) {
	if len(subnetId) == 0 {
		return nil, nil
	}

	response, err := c.vnClient.GetSubnet(ctx, core.GetSubnetRequest{
		SubnetId:        &subnetId,
		RequestMetadata: common.RequestMetadata{},
	})
	if err != nil {
		return nil, err
	}

	subnet := response.Subnet
	return &subnet, nil
}
