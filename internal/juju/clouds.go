package juju

import (
	"github.com/juju/juju/api/client/cloud"
	"github.com/juju/juju/apiserver/common"
	jujucloud "github.com/juju/juju/cloud"
	"github.com/juju/juju/rpc/params"
	"github.com/juju/names/v4"
)

type cloudsClient struct {
	ConnectionFactory
}

type ReadCloudInput struct {
	Name string
}

type CreateCloudInput struct {
	Name   string
	Params params.Cloud
}

type UpdateCloudInput struct {
	Name   string
	Params params.Cloud
}

type RemoveCloudInput struct {
	Name string
}

func newCloudsClient(cf ConnectionFactory) *cloudsClient {
	return &cloudsClient{
		ConnectionFactory: cf,
	}
}

func (c cloudsClient) ReadCloud(input ReadCloudInput) (*jujucloud.Cloud, error) {
	conn, err := c.GetConnection(nil)
	if err != nil {
		return nil, err
	}

	client := cloud.NewClient(conn)
	defer client.Close()

	cloud, err := client.Cloud(names.NewCloudTag(input.Name))
	if err != nil {
		return nil, err
	}

	authTypes := make([]string, len(cloud.AuthTypes))
	for i, authType := range cloud.AuthTypes {
		authTypes[i] = string(authType)
	}

	return &cloud, nil
}

func (c cloudsClient) CreateCloud(input CreateCloudInput) error {
	conn, err := c.GetConnection(nil)
	if err != nil {
		return err
	}

	client := cloud.NewClient(conn)
	defer client.Close()

	newCloud := common.CloudFromParams(input.Name, input.Params)
	cloudErr := client.AddCloud(newCloud, false)
	if cloudErr != nil {
		return cloudErr
	}

	return nil
}

func (c cloudsClient) UpdateCloud(input UpdateCloudInput) error {
	conn, err := c.GetConnection(nil)
	if err != nil {
		return err
	}

	client := cloud.NewClient(conn)
	defer client.Close()

	newCloud := common.CloudFromParams(input.Name, input.Params)
	cloudErr := client.UpdateCloud(newCloud)

	if cloudErr != nil {
		return cloudErr
	}

	return nil
}

func (c cloudsClient) RemoveCloud(input RemoveCloudInput) error {
	conn, err := c.GetConnection(nil)
	if err != nil {
		return err
	}

	client := cloud.NewClient(conn)
	defer client.Close()

	cloudErr := client.RemoveCloud(input.Name)
	if cloudErr != nil {
		return cloudErr
	}

	return nil
}
