// Package health provides some utility functions to health-check a server. The implementation
// is based on protobuf. Users need to write their own implementations if other IDLs are used.
package health

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health"
)

// HealthCheck is the client side function to health-check a server
func HealthCheck(t time.Duration, cc *grpc.ClientConn) error {
	ctx, _ := context.WithTimeout(context.Background(), t)
	hc := healthpb.NewHealthCheckClient(cc)
	req := new(healthpb.HealthCheckRequest)
	_, err := hc.Check(ctx, req)
	return err
}

type HealthServer struct {
}

func (s *HealthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	out := new(healthpb.HealthCheckResponse)
	return out, nil
}
