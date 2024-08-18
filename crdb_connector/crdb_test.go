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
	"errors"
	"testing"

	"github.com/fkmatsuda/dbconnector"
	"github.com/fkmatsuda/errorex"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/spf13/viper"

	"github.com/fkmatsuda/dbconnector/test"
	dbconnector_test "github.com/fkmatsuda/dbconnector/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRDBConnector(t *testing.T) {

	envTest := viper.New()
	envTest.SetEnvPrefix("CRDB_TEST")
	envTest.SetDefault("URL", "")
	envTest.AutomaticEnv()

	testServerURL := envTest.GetString("URL")

	if testServerURL == "" {
		t.Log("creates a new test server")
		ts, err := testserver.NewTestServer()
		assert.NoError(t, err)
		defer ts.Stop()
		testServerURL = ts.PGURL().String()
	} else {
		t.Log("use existing test server")
	}

	t.Logf("test server url: %v\n", testServerURL)

	tenants := []dbconnector.TenantConfig{test.NewMockTenantConfig(
		"crdbtest",
		"Test CockraochDB",
		testServerURL,
	)}

	// create a mock tenant provider
	tenantProvider := &dbconnector_test.MockTenantProvider{}
	tenantProvider.On("Configure", mock.Anything).Return(nil)
	tenantProvider.On("LoadTenants").Return(tenants, nil)

	// Test connection success
	connector, err := NewConnector(tenantProvider)
	assert.NoError(t, err)

	// create minimal test schema
	conn, err := connector.Connect(context.Background(), "crdbtest")
	assert.NoError(t, err)
	defer func(conn dbconnector.Database) {
		err := conn.Close(context.Background())
		assert.NoError(t, err)
	}(conn)

	err = conn.RunInTransaction(context.Background(), func(ctx context.Context, tx dbconnector.Transaction) error {
		_, err := tx.Exec(ctx, "create table if not exists test_table (id int, name text)")
		if err != nil {
			return err
		}
		//goland:noinspection SqlResolve
		r, err := tx.Exec(ctx, "insert into test_table (id, name) values ($1, $2) on conflict do nothing ", 1, "test 1")
		if err != nil {
			return err
		}
		ra, err := r.RowsAffected()
		if err != nil {
			return err
		}
		if ra != 1 {
			return errors.New("RowsAffected() != 1")
		}
		//goland:noinspection SqlResolve
		r, err = tx.Exec(ctx, "insert into test_table (id, name) values ($1, $2) on conflict do nothing ", 2, "test 2")
		if err != nil {
			return err
		}
		ra, err = r.RowsAffected()
		if err != nil {
			return err
		}
		if ra != 1 {
			return errors.New("RowsAffected() != 1")
		}
		//goland:noinspection SqlResolve
		r, err = tx.Exec(ctx, "insert into test_table (id, name) values ($1, $2) on conflict do nothing ", 3, "test 3")
		if err != nil {
			return err
		}
		ra, err = r.RowsAffected()
		if err != nil {
			return err
		}
		if ra != 1 {
			return errors.New("RowsAffected() != 1")
		}
		return nil
	})
	assert.NoError(t, err)

	// Test Transaction failed
	t.Run("Test Transaction failed", func(t *testing.T) {
		err = conn.RunInTransaction(context.Background(), func(ctx context.Context, tx dbconnector.Transaction) error {
			//goland:noinspection SqlResolve
			_, err := tx.Exec(ctx, "insert into error_table (id, name) values ($1, $2)", 1, "test")
			if err != nil {
				return err
			}
			return nil
		})
		assert.Error(t, err)
		ex, ok := err.(errorex.EX)
		assert.True(t, ok)
		assert.Equal(t, dbconnector.ErrCodeCannotCommitTx, ex.Code())
		detail, ok := ex.Detail().(dbconnector.DatabaseErrorDetail)
		assert.True(t, ok)
		assert.Equal(t, "42P01", detail.DatabaseErrorCode)
	})

	// Test Query failed
	t.Run("Test Query failed", func(t *testing.T) {
		//goland:noinspection SqlResolve
		_, err = conn.Query(context.Background(), "select * from error_table")
		assert.Error(t, err)
	})

	// Test Query success
	t.Run("Test Query success", func(t *testing.T) {
		connector, err := NewConnector(tenantProvider)
		if err != nil {
			t.Errorf("NewConnector() = %v", err)
		}
		conn, err := connector.Connect(context.Background(), "crdbtest")
		assert.NoError(t, err)
		defer func(conn dbconnector.Database) {
			err := conn.Close(context.Background())
			assert.NoError(t, err)
		}(conn)

		//goland:noinspection SqlResolve
		rows, err := conn.Query(context.Background(), "select * from test_table")
		assert.NoError(t, err)
		defer func(rows dbconnector.Rows) {
			err := rows.Close()
			assert.NoError(t, err)
		}(rows)
		assert.NotNil(t, rows)
		for rows.Next() {
			var id int
			var name string
			err := rows.Scan(&id, &name)
			assert.NoError(t, err)
			switch {
			case id == 1:
				assert.Equal(t, "test 1", name)
			case id == 2:
				assert.Equal(t, "test 2", name)
			case id == 3:
				assert.Equal(t, "test 3", name)
			default:
				t.Errorf("Scan() = %v; want id 1, 2 or 3", id)
			}
		}
	})

	// Test QueryRow failed
	t.Run("Test QueryRow failed", func(t *testing.T) {
		connector, err := NewConnector(tenantProvider)
		assert.NoError(t, err)
		conn, err := connector.Connect(context.Background(), "crdbtest")
		assert.NoError(t, err)
		defer func(conn dbconnector.Database) {
			err := conn.Close(context.Background())
			assert.NoError(t, err)
		}(conn)

		//goland:noinspection SqlResolve
		row := conn.QueryRow(context.Background(), "select count(*) from error_table")
		assert.NotNil(t, row)
		var count int
		err = row.Scan(&count)
		assert.Error(t, err)
	})

	// Test QueryRow success
	t.Run("Test QueryRow success", func(t *testing.T) {
		connector, err := NewConnector(tenantProvider)
		assert.NoError(t, err)
		conn, err := connector.Connect(context.Background(), "crdbtest")
		assert.NoError(t, err)
		defer func(conn dbconnector.Database) {
			err := conn.Close(context.Background())
			assert.NoError(t, err)
		}(conn)

		//goland:noinspection SqlResolve
		row := conn.QueryRow(context.Background(), "select count(*) from test_table")
		assert.NotNil(t, row)
		var count int
		err = row.Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)

	})

	// Cleanup
	database, err := connector.Connect(context.Background(), "crdbtest")
	assert.NoError(t, err)
	defer func(database dbconnector.Database) {
		err := database.Close(context.Background())
		assert.NoError(t, err)
	}(database)

	err = database.RunInTransaction(context.Background(), func(ctx context.Context, tx dbconnector.Transaction) error {
		_, err := tx.Exec(ctx, "drop table if exists test_table")
		if err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)

}
