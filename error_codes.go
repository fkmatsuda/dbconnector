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

package dbconnector

import "github.com/fkmatsuda/errorex"

const (
	ModuleCode = "dbconnector"

	ErrCodeConnectionFailed = ModuleCode + ".002"
	ErrCodeTenantNotFound   = ModuleCode + ".003"
	ErrCodeCannotBeginTx    = ModuleCode + ".005"
	ErrCodeCannotCommitTx   = ModuleCode + ".006"
	ErrCodeCannotRollbackTx = ModuleCode + ".007"
	ErrCodeNotSupported     = ModuleCode + ".008"
	ErrCodeQueryFailed      = ModuleCode + ".009"
	ErrCodeCloseFailed      = ModuleCode + ".010"
	ErrCodeGenericDBError   = ModuleCode + ".011"
)

func init() {
	errorex.RegisterErrorCode(ErrCodeConnectionFailed, "connection failed", DatabaseErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeTenantNotFound, "tenant not found", TenantErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeCannotBeginTx, "cannot begin transaction", DatabaseErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeCannotCommitTx, "cannot commit transaction", DatabaseErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeCannotRollbackTx, "cannot rollback transaction", RollbackErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeNotSupported, "not supported", TenantErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeQueryFailed, "query failed", QueryErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeCloseFailed, "close failed", DatabaseErrorDetail{})
	errorex.RegisterErrorCode(ErrCodeGenericDBError, "generic database error", DatabaseErrorDetail{})
}

// TenantErrorDetail is a struct that contains the details of an error returned by TenantError.
type TenantErrorDetail struct {
	TenantID string `json:"tenantId"`
}

// DatabaseErrorDetail is a struct that contains the details of an error returned by ConnectionError.
type DatabaseErrorDetail struct {
	TenantErrorDetail
	DatabaseErrorCode string `json:"databaseErrorCode"`
	DatabaseError     string `json:"databaseError"`
}

// QueryErrorDetail is a struct that contains the details of an error returned by QueryError.
type QueryErrorDetail struct {
	DatabaseErrorDetail
	QueryScript string        `json:"queryScript"`
	QueryArgs   []interface{} `json:"args"`
}

// RollbackErrorDetail is a struct that contains the details of an error returned by RollbackError.
type RollbackErrorDetail struct {
	DatabaseError errorex.EX `json:"databaseError"`
	OriginalError errorex.EX `json:"originalError"`
}
