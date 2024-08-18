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
	"github.com/fkmatsuda/dbconnector"

	"github.com/fkmatsuda/errorex"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgErrorConverter struct {
	errorex.BaseErrorConverter
	errCode  string
	tenantID string
}

func (c *pgErrorConverter) ConvertError(err error) errorex.EX {
	if detail, ok := pgErrorDetail(c, err); ok {
		return errorex.New(c.errCode, detail)
	}
	return c.ConvertError(err)
}

func pgErrorDetail(c *pgErrorConverter, err error) (dbconnector.DatabaseErrorDetail, bool) {
	if pgError, ok := err.(*pgconn.PgError); ok {
		detail := dbconnector.DatabaseErrorDetail{
			TenantErrorDetail: dbconnector.TenantErrorDetail{TenantID: c.tenantID},
			DatabaseError:     pgError.Message,
			DatabaseErrorCode: pgError.Code,
		}
		return detail, true
	}
	return dbconnector.DatabaseErrorDetail{}, false
}

func NewCannotCommitTxErrorConverter(tenantID string) errorex.ErrorConverter {
	return errorex.BuildErrorConverterChain(newPGErrorConverter(dbconnector.ErrCodeCannotCommitTx, tenantID))
}

func NewGenericDbErrorConverter(tenantID string) errorex.ErrorConverter {
	return errorex.BuildErrorConverterChain(newPGErrorConverter(dbconnector.ErrCodeGenericDBError, tenantID))
}

func newPGErrorConverter(errorCode, tenantID string) errorex.ErrorConverter {
	return &pgErrorConverter{
		errCode:  errorCode,
		tenantID: tenantID,
	}
}
