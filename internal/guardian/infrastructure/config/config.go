package config

type appConfig struct {
	Healthcheck  healthcheck   `hcl:"healthcheck,block"`
	UserProvider *userProvider `hcl:"userprovider,block"`
	Servers      []server      `hcl:"server,block"`
}

type healthcheck struct {
	Address string `hcl:"address"`
	Path    string `hcl:"path"`
}

const (
	ldapUserProviderType = "ldap"
)

type userProvider struct {
	Type    string `hcl:"type,label"`
	Address string `hcl:"address"`
}
