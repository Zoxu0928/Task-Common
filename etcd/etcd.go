package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	*clientv3.Client
}

func (c *Client) Close() {
	_ = c.Client.Close()
}

func NewClient(config clientv3.Config) (*Client, error) {
	c, e := clientv3.New(config)
	if e != nil {
		return nil, e
	}
	return &Client{c}, nil
}
