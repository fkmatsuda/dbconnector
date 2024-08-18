/*
 *   Copyright (c) 2024 fkmatsuda <fabio@fkmatsuda.dev>
 *   All rights reserved.

 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:

 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.

 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 */

package config

import (
	"net/url"

	"github.com/fkmatsuda/dbconnector"

	"github.com/spf13/viper"
)

// loadEnv loads the environment variables.
func loadEnv() {
	// load tenant config using viper
	vprefix := viper.New()
	vprefix.SetEnvPrefix("TENANT_PROVIDER")
	vprefix.AutomaticEnv()

	_defaultTenantProviderConfig.url = vprefix.GetString("URL")
	_defaultTenantProviderConfig.token = vprefix.GetString("TOKEN")

	_defaultDbConnectorConfig = &dbConnectorConfig{
		tenantProviderConfig: _defaultTenantProviderConfig,
	}

}

// TenantProviderConfig is the interface for tenant provider configuration.
type TenantProviderConfig interface {
	// URL is the URL of the tenant provider.
	URL() *url.URL

	// Token is the token for the tenant provider.
	Token() string
}

// DbConnectorConfig is the interface for dbconnector configuration.
type DbConnectorConfig interface {
	// TenantProviderConfig returns the tenant provider configuration.
	TenantProviderConfig() TenantProviderConfig
}

type tenantProviderConfig struct {
	url   string `mapstructure:"TENANT_PROVIDER_URL"`
	token string `mapstructure:"TENANT_PROVIDER_TOKEN"`
}

type dbConnectorConfig struct {
	tenantProviderConfig TenantProviderConfig
}

var (
	// DefaultTenantProviderConfig is the default tenant provider configuration.
	_defaultTenantProviderConfig = &tenantProviderConfig{}

	// DefaultDbConnectorConfig is the default dbconnector configuration.
	_defaultDbConnectorConfig *dbConnectorConfig
)

// CurrentConfig returns the current configuration.
func CurrentConfig() DbConnectorConfig {
	if _defaultDbConnectorConfig == nil {
		loadEnv()
	}
	return _defaultDbConnectorConfig
}

// URL implements TenantProviderConfig.
func (c *tenantProviderConfig) URL() *url.URL {
	url, err := url.Parse(c.url)
	if err != nil {
		panic(err)
	}
	return url
}

// Token implements TenantProviderConfig.
func (c *tenantProviderConfig) Token() string {
	return c.token
}

// TenantProviderConfig implements DbConnectorConfig.
func (c *dbConnectorConfig) TenantProviderConfig() TenantProviderConfig {
	return c.tenantProviderConfig
}

// simpleConfigTenantProvider is a simple tenant provider configuration.
type simpleConfigTenantProvider struct {
	url      string
	tenantID string
}

// Configure implements dbconnector.TenantProvider.
func (p *simpleConfigTenantProvider) Configure() error {
	return nil
}

// LoadTenants implements dbconnector.TenantProvider.
func (p *simpleConfigTenantProvider) LoadTenants() ([]dbconnector.TenantConfig, error) {
	tenants := []dbconnector.TenantConfig{
		dbconnector.NewTenantConfigBuilder().
			WithTenantID(p.tenantID).
			WithTenantName("__Default").
			WithDatabaseURL(p.url).Build(),
	}

	return tenants, nil

}

// NewSimpleConfigTenantProvider creates a new tenant provider with a simple configuration.
func NewSimpleConfigTenantProvider(url string, tenantID string) dbconnector.TenantProvider {
	return &simpleConfigTenantProvider{
		url:      url,
		tenantID: tenantID,
	}
}
