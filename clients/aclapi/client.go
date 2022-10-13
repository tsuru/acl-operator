package aclapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Client interface {
	AppRules(ctx context.Context, appName string) ([]Rule, error)
}

type Rule struct {
	RuleID      string
	RuleName    string
	Source      RuleType
	Destination RuleType
	Removed     bool
	Metadata    map[string]string
	Created     time.Time
	Creator     string
}

type RuleType struct {
	TsuruApp          *TsuruAppRule          `json:"TsuruApp,omitempty"`
	KubernetesService *KubernetesServiceRule `json:"KubernetesService,omitempty"`
	ExternalDNS       *ExternalDNSRule       `json:"ExternalDNS,omitempty"`
	ExternalIP        *ExternalIPRule        `json:"ExternalIP,omitempty"`
	RpaasInstance     *RpaasInstanceRule     `json:"RpaasInstance,omitempty"`
}

type ProtoPorts []ProtoPort

type ProtoPort struct {
	Protocol string
	Port     uint16
}

type TsuruAppRule struct {
	AppName  string
	PoolName string
}

type KubernetesServiceRule struct {
	Namespace   string
	ServiceName string
	ClusterName string
}

type ExternalDNSRule struct {
	Name             string
	Ports            ProtoPorts
	SyncWholeNetwork bool
}

type ExternalIPRule struct {
	IP               string
	Ports            ProtoPorts
	SyncWholeNetwork bool
}

type RpaasInstanceRule struct {
	ServiceName string
	Instance    string
}

func New(host, user, password string) Client {
	return &client{
		host:       host,
		user:       user,
		password:   password,
		httpClient: *http.DefaultClient,
	}
}

type client struct {
	host       string
	user       string
	password   string
	httpClient http.Client
}

func (c *client) AppRules(ctx context.Context, appName string) ([]Rule, error) {
	req, err := http.NewRequest(http.MethodGet, c.host+"/apps/"+appName+"/rules", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	result := []Rule{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
