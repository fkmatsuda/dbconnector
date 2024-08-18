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

package crdb_connector

import (
	"context"

	"github.com/fkmatsuda/dbconnector"
	"github.com/fkmatsuda/dbconnector/pgsql_connector"

	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
)

// CrdbConnector is the struct for the CockroachDB connector.
type CrdbConnector struct {
	pgsql_connector.PgsqlConnector
}

// NewConnector creates a new database connector.
func NewConnector(tenantProvider dbconnector.TenantProvider) (dbconnector.Connector, error) {
	// get PgsqlConnector
	connector, err := pgsql_connector.NewConnector(tenantProvider)
	if err != nil {
		return nil, err
	}
	// create crdb connector
	crdbConnector := CrdbConnector{
		PgsqlConnector: *(connector.(*pgsql_connector.PgsqlConnector)),
	}
	return &crdbConnector, nil
}

// CrdbDatabase is the struct for the CockroachDB database.
type CrdbDatabase struct {
	pgsql_connector.PgsqlDatabase
}

// override PgsqlConnector.Connect
func (c *CrdbConnector) Connect(ctx context.Context, tenantID string) (dbconnector.Database, error) {
	// get database
	database, err := c.PgsqlConnector.Connect(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// create crdb database
	crdbDatabase := CrdbDatabase{
		PgsqlDatabase: *(database.(*pgsql_connector.PgsqlDatabase)),
	}
	return &crdbDatabase, nil
}

// override PgsqlDatabase.RunInTransaction
func (d *CrdbDatabase) RunInTransaction(ctx context.Context, fn dbconnector.TransactionFN) error {
	errorConverter := pgsql_connector.NewCannotCommitTxErrorConverter(d.TenantConfig().TenantID())

	err := crdbpgx.ExecuteTx(ctx, d.PgxConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {

		dbTx := d.CreateTx(tx)

		txErr := fn(ctx, dbTx)

		return txErr

	})

	return errorConverter.ConvertError(err)
}
