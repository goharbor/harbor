/*
 *
 * Copyright 2021 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package csds implements features to dump the status (xDS responses) the
// xds_client is using.
//
// Notice: This package is EXPERIMENTAL and may be changed or removed in a later
// release.
package csds

import (
	"context"
	"fmt"
	"io"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	internalgrpclog "google.golang.org/grpc/internal/grpclog"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/xds/internal/xdsclient"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
	"google.golang.org/protobuf/types/known/timestamppb"

	v3adminpb "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	v3statusgrpc "github.com/envoyproxy/go-control-plane/envoy/service/status/v3"
	v3statuspb "github.com/envoyproxy/go-control-plane/envoy/service/status/v3"
)

var logger = grpclog.Component("xds")

const prefix = "[csds-server %p] "

func prefixLogger(s *ClientStatusDiscoveryServer) *internalgrpclog.PrefixLogger {
	return internalgrpclog.NewPrefixLogger(logger, fmt.Sprintf(prefix, s))
}

// ClientStatusDiscoveryServer provides an implementation of the Client Status
// Discovery Service (CSDS) for exposing the xDS config of a given client. See
// https://github.com/envoyproxy/envoy/blob/main/api/envoy/service/status/v3/csds.proto.
//
// For more details about the gRPC implementation of CSDS, refer to gRPC A40 at:
// https://github.com/grpc/proposal/blob/master/A40-csds-support.md.
type ClientStatusDiscoveryServer struct {
	logger *internalgrpclog.PrefixLogger

	mu             sync.Mutex
	xdsClient      xdsclient.XDSClient
	xdsClientClose func()
}

// NewClientStatusDiscoveryServer returns an implementation of the CSDS server
// that can be registered on a gRPC server.
func NewClientStatusDiscoveryServer() (*ClientStatusDiscoveryServer, error) {
	c, close, err := xdsclient.New()
	if err != nil {
		logger.Warningf("Failed to create xDS client: %v", err)
	}
	s := &ClientStatusDiscoveryServer{xdsClient: c, xdsClientClose: close}
	s.logger = prefixLogger(s)
	s.logger.Infof("Created CSDS server, with xdsClient %p", c)
	return s, nil
}

// StreamClientStatus implementations interface ClientStatusDiscoveryServiceServer.
func (s *ClientStatusDiscoveryServer) StreamClientStatus(stream v3statusgrpc.ClientStatusDiscoveryService_StreamClientStatusServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		resp, err := s.buildClientStatusRespForReq(req)
		if err != nil {
			return err
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

// FetchClientStatus implementations interface ClientStatusDiscoveryServiceServer.
func (s *ClientStatusDiscoveryServer) FetchClientStatus(_ context.Context, req *v3statuspb.ClientStatusRequest) (*v3statuspb.ClientStatusResponse, error) {
	return s.buildClientStatusRespForReq(req)
}

// buildClientStatusRespForReq fetches the status from the client, and returns
// the response to be sent back to xdsclient.
//
// If it returns an error, the error is a status error.
func (s *ClientStatusDiscoveryServer) buildClientStatusRespForReq(req *v3statuspb.ClientStatusRequest) (*v3statuspb.ClientStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.xdsClient == nil {
		return &v3statuspb.ClientStatusResponse{}, nil
	}
	// Field NodeMatchers is unsupported, by design
	// https://github.com/grpc/proposal/blob/master/A40-csds-support.md#detail-node-matching.
	if len(req.NodeMatchers) != 0 {
		return nil, status.Errorf(codes.InvalidArgument, "node_matchers are not supported, request contains node_matchers: %v", req.NodeMatchers)
	}

	dump := s.xdsClient.DumpResources()
	ret := &v3statuspb.ClientStatusResponse{
		Config: []*v3statuspb.ClientConfig{
			{
				Node:              s.xdsClient.BootstrapConfig().NodeProto,
				GenericXdsConfigs: dumpToGenericXdsConfig(dump),
			},
		},
	}
	return ret, nil
}

// Close cleans up the resources.
func (s *ClientStatusDiscoveryServer) Close() {
	if s.xdsClientClose != nil {
		s.xdsClientClose()
	}
}

func dumpToGenericXdsConfig(dump map[string]map[string]xdsresource.UpdateWithMD) []*v3statuspb.ClientConfig_GenericXdsConfig {
	var ret []*v3statuspb.ClientConfig_GenericXdsConfig
	for typeURL, updates := range dump {
		for name, update := range updates {
			config := &v3statuspb.ClientConfig_GenericXdsConfig{
				TypeUrl:      typeURL,
				Name:         name,
				VersionInfo:  update.MD.Version,
				XdsConfig:    update.Raw,
				LastUpdated:  timestamppb.New(update.MD.Timestamp),
				ClientStatus: serviceStatusToProto(update.MD.Status),
			}
			if errState := update.MD.ErrState; errState != nil {
				config.ErrorState = &v3adminpb.UpdateFailureState{
					LastUpdateAttempt: timestamppb.New(errState.Timestamp),
					Details:           errState.Err.Error(),
					VersionInfo:       errState.Version,
				}
			}
			ret = append(ret, config)
		}
	}
	return ret
}

func serviceStatusToProto(serviceStatus xdsresource.ServiceStatus) v3adminpb.ClientResourceStatus {
	switch serviceStatus {
	case xdsresource.ServiceStatusUnknown:
		return v3adminpb.ClientResourceStatus_UNKNOWN
	case xdsresource.ServiceStatusRequested:
		return v3adminpb.ClientResourceStatus_REQUESTED
	case xdsresource.ServiceStatusNotExist:
		return v3adminpb.ClientResourceStatus_DOES_NOT_EXIST
	case xdsresource.ServiceStatusACKed:
		return v3adminpb.ClientResourceStatus_ACKED
	case xdsresource.ServiceStatusNACKed:
		return v3adminpb.ClientResourceStatus_NACKED
	default:
		return v3adminpb.ClientResourceStatus_UNKNOWN
	}
}
