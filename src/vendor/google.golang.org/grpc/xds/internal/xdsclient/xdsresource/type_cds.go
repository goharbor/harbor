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
 */

package xdsresource

import (
	"time"

	"google.golang.org/protobuf/types/known/anypb"
)

// ClusterType is the type of cluster from a received CDS response.
type ClusterType int

const (
	// ClusterTypeEDS represents the EDS cluster type, which will delegate endpoint
	// discovery to the management server.
	ClusterTypeEDS ClusterType = iota
	// ClusterTypeLogicalDNS represents the Logical DNS cluster type, which essentially
	// maps to the gRPC behavior of using the DNS resolver with pick_first LB policy.
	ClusterTypeLogicalDNS
	// ClusterTypeAggregate represents the Aggregate Cluster type, which provides a
	// prioritized list of clusters to use. It is used for failover between clusters
	// with a different configuration.
	ClusterTypeAggregate
)

// ClusterLRSServerConfigType is the type of LRS server config.
type ClusterLRSServerConfigType int

const (
	// ClusterLRSOff indicates LRS is off (loads are not reported for this
	// cluster).
	ClusterLRSOff ClusterLRSServerConfigType = iota
	// ClusterLRSServerSelf indicates loads should be reported to the same
	// server (the authority) where the CDS resp is received from.
	ClusterLRSServerSelf
)

// ClusterLBPolicyRingHash represents ring_hash lb policy, and also contains its
// config.
type ClusterLBPolicyRingHash struct {
	MinimumRingSize uint64
	MaximumRingSize uint64
}

// OutlierDetection is the outlier detection configuration for a cluster.
type OutlierDetection struct {
	// Interval is the time interval between ejection analysis sweeps. This can
	// result in both new ejections as well as addresses being returned to
	// service. Defaults to 10s.
	Interval time.Duration
	// BaseEjectionTime is the base time that a host is ejected for. The real
	// time is equal to the base time multiplied by the number of times the host
	// has been ejected and is capped by MaxEjectionTime. Defaults to 30s.
	BaseEjectionTime time.Duration
	// MaxEjectionTime is the maximum time that an address is ejected for. If
	// not specified, the default value (300s) or the BaseEjectionTime value is
	// applied, whichever is larger.
	MaxEjectionTime time.Duration
	// MaxEjectionPercent is the maximum % of an upstream cluster that can be
	// ejected due to outlier detection. Defaults to 10% but will eject at least
	// one host regardless of the value.
	MaxEjectionPercent uint32
	// SuccessRateStdevFactor is used to determine the ejection threshold for
	// success rate outlier ejection. The ejection threshold is the difference
	// between the mean success rate, and the product of this factor and the
	// standard deviation of the mean success rate: mean - (stdev *
	// success_rate_stdev_factor). This factor is divided by a thousand to get a
	// double. That is, if the desired factor is 1.9, the runtime value should
	// be 1900. Defaults to 1900.
	SuccessRateStdevFactor uint32
	// EnforcingSuccessRate is the % chance that a host will be actually ejected
	// when an outlier status is detected through success rate statistics. This
	// setting can be used to disable ejection or to ramp it up slowly. Defaults
	// to 100.
	EnforcingSuccessRate uint32
	// SuccessRateMinimumHosts is the number of hosts in a cluster that must
	// have enough request volume to detect success rate outliers. If the number
	// of hosts is less than this setting, outlier detection via success rate
	// statistics is not performed for any host in the cluster. Defaults to 5.
	SuccessRateMinimumHosts uint32
	// SuccessRateRequestVolume is the minimum number of total requests that
	// must be collected in one interval (as defined by the interval duration
	// above) to include this host in success rate based outlier detection. If
	// the volume is lower than this setting, outlier detection via success rate
	// statistics is not performed for that host. Defaults to 100.
	SuccessRateRequestVolume uint32
	// FailurePercentageThreshold is the failure percentage to use when
	// determining failure percentage-based outlier detection. If the failure
	// percentage of a given host is greater than or equal to this value, it
	// will be ejected. Defaults to 85.
	FailurePercentageThreshold uint32
	// EnforcingFailurePercentage is the % chance that a host will be actually
	// ejected when an outlier status is detected through failure percentage
	// statistics. This setting can be used to disable ejection or to ramp it up
	// slowly. Defaults to 0.
	EnforcingFailurePercentage uint32
	// FailurePercentageMinimumHosts is the minimum number of hosts in a cluster
	// in order to perform failure percentage-based ejection. If the total
	// number of hosts in the cluster is less than this value, failure
	// percentage-based ejection will not be performed. Defaults to 5.
	FailurePercentageMinimumHosts uint32
	// FailurePercentageRequestVolume is the minimum number of total requests
	// that must be collected in one interval (as defined by the interval
	// duration above) to perform failure percentage-based ejection for this
	// host. If the volume is lower than this setting, failure percentage-based
	// ejection will not be performed for this host. Defaults to 50.
	FailurePercentageRequestVolume uint32
}

// ClusterUpdate contains information from a received CDS response, which is of
// interest to the registered CDS watcher.
type ClusterUpdate struct {
	ClusterType ClusterType
	// ClusterName is the clusterName being watched for through CDS.
	ClusterName string
	// EDSServiceName is an optional name for EDS. If it's not set, the balancer
	// should watch ClusterName for the EDS resources.
	EDSServiceName string
	// LRSServerConfig contains the server where the load reports should be sent
	// to. This can be change to an interface, to support other types, e.g. a
	// ServerConfig with ServerURI, creds.
	LRSServerConfig ClusterLRSServerConfigType
	// SecurityCfg contains security configuration sent by the control plane.
	SecurityCfg *SecurityConfig
	// MaxRequests for circuit breaking, if any (otherwise nil).
	MaxRequests *uint32
	// DNSHostName is used only for cluster type DNS. It's the DNS name to
	// resolve in "host:port" form
	DNSHostName string
	// PrioritizedClusterNames is used only for cluster type aggregate. It represents
	// a prioritized list of cluster names.
	PrioritizedClusterNames []string

	// LBPolicy is the lb policy for this cluster.
	//
	// This only support round_robin and ring_hash.
	// - if it's nil, the lb policy is round_robin
	// - if it's not nil, the lb policy is ring_hash, the this field has the config.
	//
	// When we add more support policies, this can be made an interface, and
	// will be set to different types based on the policy type.
	LBPolicy *ClusterLBPolicyRingHash

	// OutlierDetection is the outlier detection configuration for this cluster.
	// If nil, it means this cluster does not use the outlier detection feature.
	OutlierDetection *OutlierDetection

	// Raw is the resource from the xds response.
	Raw *anypb.Any
}

// ClusterUpdateErrTuple is a tuple with the update and error. It contains the
// results from unmarshal functions. It's used to pass unmarshal results of
// multiple resources together, e.g. in maps like `map[string]{Update,error}`.
type ClusterUpdateErrTuple struct {
	Update ClusterUpdate
	Err    error
}
