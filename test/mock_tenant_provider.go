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

package test

import (
	"github.com/fkmatsuda/dbconnector"
	"github.com/stretchr/testify/mock"
)

type MockTenantProvider struct {
	mock.Mock
}

func (m *MockTenantProvider) Configure() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTenantProvider) LoadTenants() ([]dbconnector.TenantConfig, error) {
	args := m.Called()
	return args.Get(0).([]dbconnector.TenantConfig), args.Error(1)
}

type mockTenantConfig struct {
	id   string
	name string
	url  string
}

func (m *mockTenantConfig) TenantID() string {
	return m.id
}

func (m *mockTenantConfig) TenantName() string {
	return m.name
}

func (m *mockTenantConfig) DatabaseURL() string {
	return m.url
}

func NewMockTenantConfig(id, name, url string) dbconnector.TenantConfig {
	return &mockTenantConfig{id, name, url}
}
