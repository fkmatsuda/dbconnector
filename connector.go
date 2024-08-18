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

// Package dbconnector provides interfaces and implementations for connecting to databases.
package dbconnector

import (
	"context"
)

// TenantConfig is the configuration for a tenant.
type TenantConfig interface {
	// TenantID is the ID of the tenant.
	TenantID() string
	// TenantName is the name of the tenant.
	TenantName() string
	// DatabaseURL is the URL of the database for the tenant.
	DatabaseURL() string
}

// TenantProviderConfig is the configuration for the tenant provider.
type TenantProviderConfig struct {
	// URL is the URL of the tenant provider.
	URL string
	// Token is the token for the tenant provider.
	Token string
}

// Result is the result of a query.
type Result interface {
	// LastInsertId returns the last inserted ID.
	LastInsertId() (int64, error)

	// RowsAffected returns the number of rows affected.
	RowsAffected() (int64, error)
}

// TransactionFN is the transaction function.
type TransactionFN func(ctx context.Context, tx Transaction) error

// Row is a row in the result set.
type Row interface {
	// Scan copies the columns in the current row into the values pointed at by dest.
	Scan(dest ...interface{}) error
}

// Rows is the result set.
type Rows interface {
	Row

	// Next prepares the next row for reading.
	Next() bool

	// Err returns any error that occurred while reading.
	Err() error

	// Close closes the row.
	Close() error
}

// Query is the query interface.
type Query interface {

	// TenantConfig returns the tenant config.
	TenantConfig() TenantConfig

	// Query executes a query and returns the result.
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)

	// QueryRow executes a query that is expected to return at most one row.
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
}

// Transaction is a database transaction.
type Transaction interface {
	Query

	// Exec executes a query without returning any rows.
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// TransactionFN is the transaction function.

// Database is the database interface.
type Database interface {
	Query

	// RunInTransaction executes the given function in a transaction.
	RunInTransaction(ctx context.Context, fn TransactionFN) error

	// Close closes the database.
	Close(ctx context.Context) error
}

// Connector is the database connector.
type Connector interface {
	// Connect connects to the database.
	Connect(ctx context.Context, tenantID string) (Database, error)

	// Reload reloads the tenants and reconfigure the database pool.
	Reload() error
}

// TenantProvider is the tenant provider interface.
type TenantProvider interface {
	// Configure configures the tenant provider.
	Configure() error

	// LoadTenants loads the tenants.
	LoadTenants() ([]TenantConfig, error)
}

// tenantConfigImpl is a TenantConfig implementation.
type tenantConfigImpl struct {
	id   string
	name string
	url  string
}

// TenantID implements dbconnector.TenantConfig.
func (c *tenantConfigImpl) TenantID() string {
	return c.id
}

// TenantName implements dbconnector.TenantConfig.
func (c *tenantConfigImpl) TenantName() string {
	return c.name
}

// DatabaseURL implements dbconnector.TenantConfig.
func (c *tenantConfigImpl) DatabaseURL() string {
	return c.url
}

// TeanantConfigBuilder is a builder for TenantConfig.
type TenantConfigBuilder interface {
	// WithTenantID sets the tenant ID.
	WithTenantID(id string) TenantConfigBuilder
	// WithTenantName sets the tenant name.
	WithTenantName(name string) TenantConfigBuilder
	// WithDatabaseURL sets the database URL.
	WithDatabaseURL(url string) TenantConfigBuilder
	// Build creates a TenantConfig.
	Build() TenantConfig
}

type tenantConfigBuilder struct {
	tenantConfig tenantConfigImpl
}

// WithTenantID implements TenantConfigBuilder.
func (b *tenantConfigBuilder) WithTenantID(id string) TenantConfigBuilder {
	b.tenantConfig.id = id
	return b
}

// WithTenantName implements TenantConfigBuilder.
func (b *tenantConfigBuilder) WithTenantName(name string) TenantConfigBuilder {
	b.tenantConfig.name = name
	return b
}

// WithDatabaseURL implements TenantConfigBuilder.
func (b *tenantConfigBuilder) WithDatabaseURL(url string) TenantConfigBuilder {
	b.tenantConfig.url = url
	return b
}

// Build implements TenantConfigBuilder.
func (b *tenantConfigBuilder) Build() TenantConfig {
	return &b.tenantConfig
}

// NewTenantConfigBuilder creates a new TenantConfigBuilder.
func NewTenantConfigBuilder() TenantConfigBuilder {
	return &tenantConfigBuilder{
		tenantConfig: tenantConfigImpl{},
	}
}
