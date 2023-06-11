// Copyright 2018 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpcflags

import (
	"flag"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// tlsCertFile is the flag-assigned value for the path to the Trillian server's TLS certificate.
var tlsCertFile = flag.String("tls_cert_file", "", "Path to the file containing the Trillian server's PEM-encoded public TLS certificate. If unset, unsecured connections will be used")

// NewClientDialOptionsFromFlags returns a list of grpc.DialOption values to be
// passed as DialOption arguments to grpc.Dial
func NewClientDialOptionsFromFlags() ([]grpc.DialOption, error) {
	dialOpts := []grpc.DialOption{}

	if *tlsCertFile == "" {
		glog.Warning("Using an insecure gRPC connection to Trillian")
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		creds, err := credentials.NewClientTLSFromFile(*tlsCertFile, "")
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}

	return dialOpts, nil
}
