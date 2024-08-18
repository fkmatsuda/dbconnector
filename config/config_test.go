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

package config_test

import (
	"testing"

	"github.com/fkmatsuda/dbconnector/config"

	"github.com/stretchr/testify/assert"
)

func TestCurrentConfig(t *testing.T) {

	t.Setenv("TENANT_PROVIDER_URL", "http://test.tenantprovider.com")
	t.Setenv("TENANT_PROVIDER_TOKEN", "token")

	t.Run("Test CurrentConfig", func(t *testing.T) {
		config := config.CurrentConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "http://test.tenantprovider.com", config.TenantProviderConfig().URL().String())
		assert.Equal(t, "token", config.TenantProviderConfig().Token())
	})
}

func TestSimpleConfigTenantProviderConfig(t *testing.T) {

	const databaseUrl = "pgsql://localhost:5432/pgtest?sslmode=disable"
	const tenantId = "simple"

	t.Run("Test SimpleConfigTenantProviderConfig", func(t *testing.T) {
		provider := config.NewSimpleConfigTenantProvider(databaseUrl, tenantId)
		assert.NotNil(t, provider)
		tenants, err := provider.LoadTenants()
		assert.NoError(t, err)
		assert.NotNil(t, tenants)
		assert.Equal(t, 1, len(tenants))
		assert.Equal(t, tenantId, tenants[0].TenantID())
		assert.Equal(t, databaseUrl, tenants[0].DatabaseURL())
		assert.Equal(t, "__Default", tenants[0].TenantName())
	})
}
