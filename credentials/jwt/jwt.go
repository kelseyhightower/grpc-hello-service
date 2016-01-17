// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jwt

import (
	"io/ioutil"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type jwt struct {
	token string
}

func NewFromTokenFile(token string) (credentials.Credentials, error) {
	data, err := ioutil.ReadFile(token)
	if err != nil {
		return jwt{}, err
	}
	return jwt{string(data)}, nil
}

func (j jwt) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": j.token,
	}, nil
}

func (j jwt) RequireTransportSecurity() bool {
	return true
}
