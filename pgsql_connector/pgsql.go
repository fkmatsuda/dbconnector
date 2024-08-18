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

package pgsql_connector

import (
	"context"

	"github.com/fkmatsuda/dbconnector"

	"github.com/fkmatsuda/errorex"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgsqlConnector is the struct for the PostgreSQL connector.
type PgsqlConnector struct {
	tenantProvider        dbconnector.TenantProvider
	tenantsConfig         []dbconnector.TenantConfig
	tenantsConfigIndexMap map[string]int
}

// NewConnector creates a new database connector.
func NewConnector(tenantProvider dbconnector.TenantProvider) (dbconnector.Connector, error) {

	// configure tenant provider
	err := tenantProvider.Configure()
	if err != nil {
		return nil, err
	}

	// fetch tenants
	tenantsConfig, err := tenantProvider.LoadTenants()
	if err != nil {
		return nil, err
	}

	// create connector
	connector := PgsqlConnector{
		tenantProvider:        tenantProvider,
		tenantsConfig:         tenantsConfig,
		tenantsConfigIndexMap: make(map[string]int),
	}

	// load tenants
	err = (&connector).Reload()
	if err != nil {
		return nil, err
	}

	return &connector, nil
}

// a pointer to PgsqlConnector must implement Connector

// Connect connects to the database.
func (c *PgsqlConnector) Connect(ctx context.Context, tenantID string) (dbconnector.Database, error) {
	// get tenant index
	tenantIndex, ok := c.tenantsConfigIndexMap[tenantID]
	if !ok {
		return nil, errorex.New(dbconnector.ErrCodeTenantNotFound, dbconnector.TenantErrorDetail{TenantID: tenantID})
	}
	// get tenant config
	tenantConfig := c.tenantsConfig[tenantIndex]
	// obter o database
	database, err := c.connect(ctx, tenantConfig)
	if err != nil {
		return nil, err
	}
	return database, nil
}

func (c *PgsqlConnector) connect(ctx context.Context, config dbconnector.TenantConfig) (*PgsqlDatabase, error) {
	// get database url
	databaseURL := config.DatabaseURL()

	// connect to the database
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, errorex.New(dbconnector.ErrCodeConnectionFailed, dbconnector.DatabaseErrorDetail{
			TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: config.TenantID()},
			DatabaseError:     err.Error(),
		})
	}
	// fill the database struct
	database := PgsqlDatabase{
		config:    config,
		connector: c,
		conn:      conn,
	}
	return &database, nil
}

// Reload reloads the tenants and reconfigure the database pool.
func (c *PgsqlConnector) Reload() error {
	// load tenants
	tenantsConfig, err := c.tenantProvider.LoadTenants()
	if err != nil {
		return err
	}

	c.tenantsConfigIndexMap = make(map[string]int)

	// fill the index map
	for idx, tenantConfig := range tenantsConfig {
		c.tenantsConfigIndexMap[tenantConfig.TenantID()] = idx
	}

	return nil
}

// PgsqlDatabase is the struct for the PostgreSQL database.
type PgsqlDatabase struct {
	config    dbconnector.TenantConfig
	connector *PgsqlConnector
	conn      *pgx.Conn
}

// TenantConfig returns the tenant config.
func (p *PgsqlDatabase) TenantConfig() dbconnector.TenantConfig {
	return p.config
}

// PgxConn returns the pgx connection.
func (p *PgsqlDatabase) PgxConn() *pgx.Conn {
	return p.conn
}

func (p *PgsqlDatabase) Close(ctx context.Context) error {
	err := p.conn.Close(ctx)
	if err == nil {
		return nil
	}
	return errorex.New(dbconnector.ErrCodeCloseFailed, dbconnector.DatabaseErrorDetail{
		TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: p.TenantConfig().TenantID()},
		DatabaseError:     err.Error(),
	})
}

// Query executes a query.
func (p *PgsqlDatabase) Query(ctx context.Context, query string, args ...interface{}) (dbconnector.Rows, error) {
	// executar a query
	rows, err := p.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, errorex.New(dbconnector.ErrCodeQueryFailed,
			dbconnector.QueryErrorDetail{
				DatabaseErrorDetail: dbconnector.DatabaseErrorDetail{
					TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: p.TenantConfig().TenantID()},
					DatabaseError:     err.Error(),
				},
				QueryScript: query,
				QueryArgs:   args,
			},
		)
	}
	return &pgsqlRows{
		rows: rows,
	}, nil
}

// QueryRow executes a query and returns a row.
func (p *PgsqlDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) dbconnector.Row {
	// executar a query
	return p.conn.QueryRow(ctx, query, args...)
}

// RunInTransaction runs a function in a transaction.
func (p *PgsqlDatabase) RunInTransaction(ctx context.Context, fn dbconnector.TransactionFN) error {
	// create a pgx transaction
	pgxTx, err := p.conn.Begin(ctx)
	if err != nil {
		return errorex.New(dbconnector.ErrCodeCannotBeginTx,
			dbconnector.DatabaseErrorDetail{
				TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: p.TenantConfig().TenantID()},
				DatabaseError:     err.Error(),
			},
		)
	}

	// fill the transaction
	tx := p.CreateTx(pgxTx)

	// run the function
	err = fn(ctx, tx)

	return tx.CommitOrRollback(ctx, err)
}

func (p *PgsqlDatabase) CreateTx(pgxTx pgx.Tx) *PgsqlTransaction {
	tx := PgsqlTransaction{
		database: p,
		tx:       pgxTx,
	}
	return &tx
}

// PgsqlResult is the struct for the PostgreSQL result.
type PgsqlResult struct {
	database *PgsqlDatabase
	result   pgconn.CommandTag
}

// LastInsertId returns the last insert id.
func (p *PgsqlResult) LastInsertId() (int64, error) {
	return 0, errorex.New(dbconnector.ErrCodeNotSupported, dbconnector.TenantErrorDetail{
		TenantID: p.database.TenantConfig().TenantID(),
	})
}

// RowsAffected returns the number of rows affected.
func (p *PgsqlResult) RowsAffected() (int64, error) {
	return p.result.RowsAffected(), nil
}

// PgsqlTransaction is the struct for the PostgreSQL transaction.
type PgsqlTransaction struct {
	database *PgsqlDatabase
	tx       pgx.Tx
}

// TenantConfig returns the tenant config.
func (p *PgsqlTransaction) TenantConfig() dbconnector.TenantConfig {
	return p.database.TenantConfig()
}

// Query executes a query.
func (p *PgsqlTransaction) Query(ctx context.Context, query string, args ...interface{}) (dbconnector.Rows, error) {
	// execute the query
	pgRow, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, errorex.New(dbconnector.ErrCodeQueryFailed,
			dbconnector.QueryErrorDetail{
				DatabaseErrorDetail: dbconnector.DatabaseErrorDetail{
					TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: p.database.TenantConfig().TenantID()},
					DatabaseError:     err.Error(),
				},
				QueryScript: query,
				QueryArgs:   args,
			},
		)
	}
	return &pgsqlRows{
		rows: pgRow,
	}, nil
}

// QueryRow executes a query and returns a row.
func (p *PgsqlTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) dbconnector.Row {
	// execute the query
	return p.tx.QueryRow(ctx, query, args...)
}

// Exec executes a query.
func (p *PgsqlTransaction) Exec(ctx context.Context, query string, args ...interface{}) (dbconnector.Result, error) {
	// executar a query
	result, err := p.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgsqlResult{
		result: result,
	}, nil
}

// CommitOrRollback commits or rollback the transaction based on the error.
func (p *PgsqlTransaction) CommitOrRollback(ctx context.Context, err error) error {
	errorConverter := NewCannotCommitTxErrorConverter(p.database.TenantConfig().TenantID())
	if err != nil {
		errEX := errorConverter.ConvertError(err)
		// rollback the transaction
		errRollback := p.tx.Rollback(ctx)
		if errRollback != nil {
			return errorex.New(dbconnector.ErrCodeCannotRollbackTx, dbconnector.RollbackErrorDetail{
				DatabaseError: errorConverter.ConvertError(errRollback),
				OriginalError: errEX,
			})
		}
		return errEX

	}
	// commit the transaction
	errCommit := p.tx.Commit(ctx)
	if err != nil {
		return errorConverter.ConvertError(errCommit)
	}
	return nil
}

type pgsqlRows struct {
	rows pgx.Rows
}

func (p *pgsqlRows) Scan(dest ...interface{}) error {
	return p.rows.Scan(dest...)
}

func (p *pgsqlRows) Next() bool {
	return p.rows.Next()
}

func (p *pgsqlRows) Err() error {
	return p.rows.Err()
}

func (p *pgsqlRows) Close() error {
	p.rows.Close()
	return nil
}
