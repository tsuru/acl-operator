package tsuruapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tsuru/tsuru/app"
)

type Client interface {
	AppInfo(ctx context.Context, appName string) (*app.App, error)
	ServiceInstanceInfo(ctx context.Context, serviceName, instance string) (*ServiceInstanceInfo, error)
}

type ServiceInstanceInfo struct {
	Pool       string
	CustomInfo map[string]interface{}
}

func New(host, token string) Client {
	return &client{
		host:  host,
		token: token,
	}
}

type client struct {
	host  string
	token string
}

func (c *client) AppInfo(ctx context.Context, appName string) (*app.App, error) {
	// TODO add cache
	var appData app.App

	req, err := http.NewRequest(http.MethodGet, c.host+"/apps/"+appName, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to request, status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&appData)
	if err != nil {
		return nil, err
	}
	if appData.Pool == "" || appData.Name == "" {
		return nil, fmt.Errorf("empty data for app %q", appName)
	}

	return &appData, nil
}

func (c *client) ServiceInstanceInfo(ctx context.Context, serviceName, instance string) (*ServiceInstanceInfo, error) {
	// TODO add cache
	info := &ServiceInstanceInfo{}
	req, err := http.NewRequest(http.MethodGet, c.host+"/services/"+serviceName+"/instances/"+instance, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to request, status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
