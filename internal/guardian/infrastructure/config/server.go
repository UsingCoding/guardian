package config

import (
	"github.com/hashicorp/hcl/v2"
)

type server struct {
	Address string `hcl:"address,label"`

	Limit limit `hcl:"limit,block"`

	Downstream []downstream `hcl:"downstream,block"`
	Upstream   []upstream   `hcl:"upstream,block"`
}

type limit struct {
	RPS   int `hcl:"rps"`
	Burst int `hcl:"burst"`
}

type downstream struct {
	ID         string                `hcl:"id,label"`
	UpstreamID string                `hcl:"upstream"`
	Rules      []rule                `hcl:"rule,block"`
	Authorizer *downstreamAuthorizer `hcl:"authorizer,block"`
}

const (
	hostRuleType       = "host"
	pathPrefixRuleType = "path-prefix"
)

type rule struct {
	Type    string   `hcl:"type,label"`
	Payload hcl.Body `hcl:",remain"`
}

type hostRule struct {
	Host string `hcl:"host"`
}

type pathPrefixRule struct {
	Path string `hcl:"path"`
}

type downstreamAuthorizer struct {
	Type    string   `hcl:"type,label"`
	Payload hcl.Body `hcl:",remain"`
}

type cookieDownstreamAuthorizer struct {
	Key string `hcl:"key"`
}

const (
	cookieDownstreamAuthorizerType = "cookie"
)

type upstream struct {
	ID         string              `hcl:"id,label"`
	Address    string              `hcl:"address"`
	Authorizer *upstreamAuthorizer `hcl:"authorizer,block"`
}

type upstreamAuthorizer struct {
	Type    string   `hcl:"type,label"`
	Payload hcl.Body `hcl:",remain"`
}

const (
	headerUpstreamAuthorizerType = "header"
)

type headerUpstreamAuthorizer struct {
	UserID   string `hcl:"userID"`
	Username string `hcl:"username"`
}
