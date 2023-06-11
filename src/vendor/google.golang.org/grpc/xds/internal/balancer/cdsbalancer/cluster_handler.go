/*
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

package cdsbalancer

import (
	"errors"
	"sync"

	"google.golang.org/grpc/xds/internal/xdsclient"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

const maxDepth = 16

var (
	errNotReceivedUpdate = errors.New("tried to construct a cluster update on a cluster that has not received an update")
	errExceedsMaxDepth   = errors.New("aggregate cluster graph exceeds max depth")
)

// clusterHandlerUpdate wraps the information received from the registered CDS
// watcher. A non-nil error is propagated to the underlying cluster_resolver
// balancer. A valid update results in creating a new cluster_resolver balancer
// (if one doesn't already exist) and pushing the update to it.
type clusterHandlerUpdate struct {
	// securityCfg is the Security Config from the top (root) cluster.
	securityCfg *xdsresource.SecurityConfig
	// lbPolicy is the lb policy from the top (root) cluster.
	//
	// Currently, we only support roundrobin or ringhash, and since roundrobin
	// does need configs, this is only set to the ringhash config, if the policy
	// is ringhash. In the future, if we support more policies, we can make this
	// an interface, and set it to config of the other policies.
	lbPolicy *xdsresource.ClusterLBPolicyRingHash

	// updates is a list of ClusterUpdates from all the leaf clusters.
	updates []xdsresource.ClusterUpdate
	err     error
}

// clusterHandler will be given a name representing a cluster. It will then
// update the CDS policy constantly with a list of Clusters to pass down to
// XdsClusterResolverLoadBalancingPolicyConfig in a stream like fashion.
type clusterHandler struct {
	parent *cdsBalancer

	// A mutex to protect entire tree of clusters.
	clusterMutex    sync.Mutex
	rootClusterName string

	createdClusters map[string]*clusterNode

	// A way to ping CDS Balancer about any updates or errors to a Node in the
	// tree. This will either get called from this handler constructing an
	// update or from a child with an error. Capacity of one as the only update
	// CDS Balancer cares about is the most recent update.
	updateChannel chan clusterHandlerUpdate
}

func newClusterHandler(parent *cdsBalancer) *clusterHandler {
	return &clusterHandler{
		parent:          parent,
		updateChannel:   make(chan clusterHandlerUpdate, 1),
		createdClusters: make(map[string]*clusterNode),
	}
}

func (ch *clusterHandler) updateRootCluster(rootClusterName string) {
	ch.clusterMutex.Lock()
	defer ch.clusterMutex.Unlock()
	if ch.createdClusters[ch.rootClusterName] == nil {
		// Construct a root node on first update.
		createClusterNode(rootClusterName, ch.parent.xdsClient, ch, 0)
		ch.rootClusterName = rootClusterName
		return
	}
	// Check if root cluster was changed. If it was, delete old one and start
	// new one, if not do nothing.
	if rootClusterName != ch.rootClusterName {
		ch.createdClusters[ch.rootClusterName].delete()
		createClusterNode(rootClusterName, ch.parent.xdsClient, ch, 0)
		ch.rootClusterName = rootClusterName
	}
}

// This function tries to construct a cluster update to send to CDS.
func (ch *clusterHandler) constructClusterUpdate() {
	if ch.createdClusters[ch.rootClusterName] == nil {
		// If root is nil, this handler is closed, ignore the update.
		return
	}
	clusterUpdate, err := ch.createdClusters[ch.rootClusterName].constructClusterUpdate(make(map[string]bool))
	if err != nil {
		// If there was an error received no op, as this can mean one of the
		// children hasn't received an update yet, or the graph continued to
		// stay in an error state. If the graph continues to stay in an error
		// state, no new error needs to be written to the update buffer as that
		// would be redundant information.
		return
	}
	if clusterUpdate == nil {
		// This means that there was an aggregated cluster with no EDS or DNS as
		// leaf nodes. No update to be written.
		return
	}
	// For a ClusterUpdate, the only update CDS cares about is the most
	// recent one, so opportunistically drain the update channel before
	// sending the new update.
	select {
	case <-ch.updateChannel:
	default:
	}
	ch.updateChannel <- clusterHandlerUpdate{
		securityCfg: ch.createdClusters[ch.rootClusterName].clusterUpdate.SecurityCfg,
		lbPolicy:    ch.createdClusters[ch.rootClusterName].clusterUpdate.LBPolicy,
		updates:     clusterUpdate,
	}
}

// close() is meant to be called by CDS when the CDS balancer is closed, and it
// cancels the watches for every cluster in the cluster tree.
func (ch *clusterHandler) close() {
	ch.clusterMutex.Lock()
	defer ch.clusterMutex.Unlock()
	if ch.createdClusters[ch.rootClusterName] == nil {
		return
	}
	ch.createdClusters[ch.rootClusterName].delete()
	ch.rootClusterName = ""
}

// This logically represents a cluster. This handles all the logic for starting
// and stopping a cluster watch, handling any updates, and constructing a list
// recursively for the ClusterHandler.
type clusterNode struct {
	// A way to cancel the watch for the cluster.
	cancelFunc func()

	// A list of children, as the Node can be an aggregate Cluster.
	children []string

	// A ClusterUpdate in order to build a list of cluster updates for CDS to
	// send down to child XdsClusterResolverLoadBalancingPolicy.
	clusterUpdate xdsresource.ClusterUpdate

	// This boolean determines whether this Node has received an update or not.
	// This isn't the best practice, but this will protect a list of Cluster
	// Updates from being constructed if a cluster in the tree has not received
	// an update yet.
	receivedUpdate bool

	clusterHandler *clusterHandler

	depth    int32
	refCount int32

	// maxDepthErr is set if this cluster node is an aggregate cluster and has a
	// child that causes the graph to exceed the maximum depth allowed. This is
	// used to show a cluster graph as being in an error state when it constructs
	// a cluster update.
	maxDepthErr error
}

// CreateClusterNode creates a cluster node from a given clusterName. This will
// also start the watch for that cluster.
func createClusterNode(clusterName string, xdsClient xdsclient.XDSClient, topLevelHandler *clusterHandler, depth int32) {
	// If the cluster has already been created, simply return, which ignores
	// duplicates.
	if topLevelHandler.createdClusters[clusterName] != nil {
		topLevelHandler.createdClusters[clusterName].refCount++
		return
	}
	c := &clusterNode{
		clusterHandler: topLevelHandler,
		depth:          depth,
		refCount:       1,
	}
	// Communicate with the xds client here.
	topLevelHandler.parent.logger.Infof("CDS watch started on %v", clusterName)
	cancel := xdsClient.WatchCluster(clusterName, c.handleResp)
	c.cancelFunc = func() {
		topLevelHandler.parent.logger.Infof("CDS watch canceled on %v", clusterName)
		cancel()
	}
	topLevelHandler.createdClusters[clusterName] = c
}

// This function cancels the cluster watch on the cluster and all of it's
// children.
func (c *clusterNode) delete() {
	c.refCount--
	if c.refCount == 0 {
		c.cancelFunc()
		delete(c.clusterHandler.createdClusters, c.clusterUpdate.ClusterName)
		for _, child := range c.children {
			if c.clusterHandler.createdClusters[child] != nil {
				c.clusterHandler.createdClusters[child].delete()
			}
		}
	}
}

// Construct cluster update (potentially a list of ClusterUpdates) for a node.
func (c *clusterNode) constructClusterUpdate(clustersSeen map[string]bool) ([]xdsresource.ClusterUpdate, error) {
	// If the cluster has not yet received an update, the cluster update is not
	// yet ready.
	if !c.receivedUpdate {
		return nil, errNotReceivedUpdate
	}
	if c.maxDepthErr != nil {
		return nil, c.maxDepthErr
	}
	// Ignore duplicates. It's ok to ignore duplicates because the second
	// occurrence of a cluster will never be used. I.e. in [C, D, C], the second
	// C will never be used (the only way to fall back to lower priority D is if
	// C is down, which means second C will never be chosen). Thus, [C, D, C] is
	// logically equivalent to [C, D].
	if clustersSeen[c.clusterUpdate.ClusterName] {
		return []xdsresource.ClusterUpdate{}, nil
	}
	clustersSeen[c.clusterUpdate.ClusterName] = true

	// Base case - LogicalDNS or EDS. Both of these cluster types will be tied
	// to a single ClusterUpdate.
	if c.clusterUpdate.ClusterType != xdsresource.ClusterTypeAggregate {
		return []xdsresource.ClusterUpdate{c.clusterUpdate}, nil
	}

	// If an aggregate construct a list by recursively calling down to all of
	// it's children.
	var childrenUpdates []xdsresource.ClusterUpdate
	for _, child := range c.children {
		childUpdateList, err := c.clusterHandler.createdClusters[child].constructClusterUpdate(clustersSeen)
		if err != nil {
			return nil, err
		}
		childrenUpdates = append(childrenUpdates, childUpdateList...)
	}
	return childrenUpdates, nil
}

// handleResp handles a xds response for a particular cluster. This function
// also handles any logic with regards to any child state that may have changed.
// At the end of the handleResp(), the clusterUpdate will be pinged in certain
// situations to try and construct an update to send back to CDS.
func (c *clusterNode) handleResp(clusterUpdate xdsresource.ClusterUpdate, err error) {
	c.clusterHandler.clusterMutex.Lock()
	defer c.clusterHandler.clusterMutex.Unlock()
	if err != nil { // Write this error for run() to pick up in CDS LB policy.
		// For a ClusterUpdate, the only update CDS cares about is the most
		// recent one, so opportunistically drain the update channel before
		// sending the new update.
		select {
		case <-c.clusterHandler.updateChannel:
		default:
		}
		c.clusterHandler.updateChannel <- clusterHandlerUpdate{err: err}
		c.receivedUpdate = false
		c.maxDepthErr = nil
		return
	}

	c.receivedUpdate = true
	c.clusterUpdate = clusterUpdate

	// If the cluster was a leaf node, if the cluster update received had change
	// in the cluster update then the overall cluster update would change and
	// there is a possibility for the overall update to build so ping cluster
	// handler to return. Also, if there was any children from previously,
	// delete the children, as the cluster type is no longer an aggregate
	// cluster.
	if clusterUpdate.ClusterType != xdsresource.ClusterTypeAggregate {
		for _, child := range c.children {
			c.clusterHandler.createdClusters[child].delete()
		}
		c.children = nil
		c.maxDepthErr = nil
		// This is an update in the one leaf node, should try to send an update
		// to the parent CDS balancer.
		//
		// Note that this update might be a duplicate from the previous one.
		// Because the update contains not only the cluster name to watch, but
		// also the extra fields (e.g. security config). There's no good way to
		// compare all the fields.
		c.clusterHandler.constructClusterUpdate()
		return
	}

	// Aggregate cluster handling.
	if len(clusterUpdate.PrioritizedClusterNames) >= 1 {
		if c.depth == maxDepth-1 {
			// For a ClusterUpdate, the only update CDS cares about is the most
			// recent one, so opportunistically drain the update channel before
			// sending the new update.
			select {
			case <-c.clusterHandler.updateChannel:
			default:
			}
			c.clusterHandler.updateChannel <- clusterHandlerUpdate{err: errExceedsMaxDepth}
			c.children = []string{}
			c.maxDepthErr = errExceedsMaxDepth
			return
		}
	}

	newChildren := make(map[string]bool)
	for _, childName := range clusterUpdate.PrioritizedClusterNames {
		newChildren[childName] = true
	}

	// These booleans help determine whether this callback will ping the overall
	// clusterHandler to try and construct an update to send back to CDS. This
	// will be determined by whether there would be a change in the overall
	// clusterUpdate for the whole tree (ex. change in clusterUpdate for current
	// cluster or a deleted child) and also if there's even a possibility for
	// the update to build (ex. if a child is created and a watch is started,
	// that child hasn't received an update yet due to the mutex lock on this
	// callback).
	var createdChild bool

	// This map will represent the current children of the cluster. It will be
	// first added to in order to represent the new children. It will then have
	// any children deleted that are no longer present.
	mapCurrentChildren := make(map[string]bool)
	for _, child := range c.children {
		mapCurrentChildren[child] = true
	}

	// Add and construct any new child nodes.
	for child := range newChildren {
		if _, inChildrenAlready := mapCurrentChildren[child]; !inChildrenAlready {
			createClusterNode(child, c.clusterHandler.parent.xdsClient, c.clusterHandler, c.depth+1)
		}
	}

	// Delete any child nodes no longer in the aggregate cluster's children.
	for child := range mapCurrentChildren {
		if _, stillAChild := newChildren[child]; !stillAChild {
			c.clusterHandler.createdClusters[child].delete()
			delete(mapCurrentChildren, child)
		}
	}

	c.children = clusterUpdate.PrioritizedClusterNames

	c.maxDepthErr = nil
	// If the cluster is an aggregate cluster, if this callback created any new
	// child cluster nodes, then there's no possibility for a full cluster
	// update to successfully build, as those created children will not have
	// received an update yet. Even if this update did not delete a child, there
	// is still a possibility for the cluster update to build, as the aggregate
	// cluster can ignore duplicated children and thus the update can fill out
	// the full cluster update tree.
	if !createdChild {
		c.clusterHandler.constructClusterUpdate()
	}
}
