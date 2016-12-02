package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultClientTimeout = 10 * time.Second
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return NewClientWithTimeout(defaultClientTimeout)
}

func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) fetch(url string) ([]byte, error) {
	res, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Scrape: error fetching %s: %v", url, err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Scrape: error reading %s: %v", url, err)
	}
	return body, nil
}

func (c *Client) fetchHtml(url string) (string, error) {
	body, err := c.fetch(url)
	if err != nil {
		return "", err
	}
	return string(body), err
}

func (c *Client) fetchJson(url string, v interface{}) error {
	body, err := c.fetch(url)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
