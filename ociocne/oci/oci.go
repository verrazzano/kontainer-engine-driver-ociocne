// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package oci

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

type Client interface {
	GetSubnetById(context.Context, string) (*core.Subnet, error)
	GetImageOCIDByName(context.Context, string, string) (string, error)
}

type ClientImpl struct {
	vnClient              core.VirtualNetworkClient
	computeClient         core.ComputeClient
	configurationProvider common.ConfigurationProvider
}

func NewClient(provider common.ConfigurationProvider) (Client, error) {
	net, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	compute, err := core.NewComputeClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	return &ClientImpl{
		vnClient:      net,
		computeClient: compute,
	}, nil
}

func (c *ClientImpl) GetImageOCIDByName(ctx context.Context, imageName, compartmentId string) (string, error) {
	images, err := c.computeClient.ListImages(ctx, core.ListImagesRequest{
		CompartmentId: &compartmentId,
		DisplayName:   &imageName,
	})
	if err != nil {
		return "", err
	}
	if len(images.Items) < 1 {
		return "", fmt.Errorf("no images found for %s/%s", compartmentId, imageName)
	}
	return *images.Items[0].Id, nil
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
