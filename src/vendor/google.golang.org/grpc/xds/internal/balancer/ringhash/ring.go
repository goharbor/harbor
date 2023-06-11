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

package ringhash

import (
	"math"
	"sort"
	"strconv"

	xxhash "github.com/cespare/xxhash/v2"
	"google.golang.org/grpc/internal/grpclog"
	"google.golang.org/grpc/resolver"
)

type ring struct {
	items []*ringEntry
}

type subConnWithWeight struct {
	sc     *subConn
	weight float64
}

type ringEntry struct {
	idx  int
	hash uint64
	sc   *subConn
}

// newRing creates a ring from the subConns stored in the AddressMap. The ring
// size is limited by the passed in max/min.
//
// ring entries will be created for each subConn, and subConn with high weight
// (specified by the address) may have multiple entries.
//
// For example, for subConns with weights {a:3, b:3, c:4}, a generated ring of
// size 10 could be:
// - {idx:0 hash:3689675255460411075  b}
// - {idx:1 hash:4262906501694543955  c}
// - {idx:2 hash:5712155492001633497  c}
// - {idx:3 hash:8050519350657643659  b}
// - {idx:4 hash:8723022065838381142  b}
// - {idx:5 hash:11532782514799973195 a}
// - {idx:6 hash:13157034721563383607 c}
// - {idx:7 hash:14468677667651225770 c}
// - {idx:8 hash:17336016884672388720 a}
// - {idx:9 hash:18151002094784932496 a}
//
// To pick from a ring, a binary search will be done for the given target hash,
// and first item with hash >= given hash will be returned.
//
// Must be called with a non-empty subConns map.
func newRing(subConns *resolver.AddressMap, minRingSize, maxRingSize uint64, logger *grpclog.PrefixLogger) *ring {
	logger.Debugf("newRing: number of subConns is %d, minRingSize is %d, maxRingSize is %d", subConns.Len(), minRingSize, maxRingSize)

	// https://github.com/envoyproxy/envoy/blob/765c970f06a4c962961a0e03a467e165b276d50f/source/common/upstream/ring_hash_lb.cc#L114
	normalizedWeights, minWeight := normalizeWeights(subConns)
	logger.Debugf("newRing: normalized subConn weights is %v", normalizedWeights)

	// Normalized weights for {3,3,4} is {0.3,0.3,0.4}.

	// Scale up the size of the ring such that the least-weighted host gets a
	// whole number of hashes on the ring.
	//
	// Note that size is limited by the input max/min.
	scale := math.Min(math.Ceil(minWeight*float64(minRingSize))/minWeight, float64(maxRingSize))
	ringSize := math.Ceil(scale)
	items := make([]*ringEntry, 0, int(ringSize))
	logger.Debugf("newRing: creating new ring of size %v", ringSize)

	// For each entry, scale*weight nodes are generated in the ring.
	//
	// Not all of these are whole numbers. E.g. for weights {a:3,b:3,c:4}, if
	// ring size is 7, scale is 6.66. The numbers of nodes will be
	// {a,a,b,b,c,c,c}.
	//
	// A hash is generated for each item, and later the results will be sorted
	// based on the hash.
	var currentHashes, targetHashes float64
	for _, scw := range normalizedWeights {
		targetHashes += scale * scw.weight
		// This index ensures that ring entries corresponding to the same
		// address hash to different values. And since this index is
		// per-address, these entries hash to the same value across address
		// updates.
		idx := 0
		for currentHashes < targetHashes {
			h := xxhash.Sum64String(scw.sc.addr + "_" + strconv.Itoa(idx))
			items = append(items, &ringEntry{hash: h, sc: scw.sc})
			idx++
			currentHashes++
		}
	}

	// Sort items based on hash, to prepare for binary search.
	sort.Slice(items, func(i, j int) bool { return items[i].hash < items[j].hash })
	for i, ii := range items {
		ii.idx = i
	}
	return &ring{items: items}
}

// normalizeWeights divides all the weights by the sum, so that the total weight
// is 1.
//
// Must be called with a non-empty subConns map.
func normalizeWeights(subConns *resolver.AddressMap) ([]subConnWithWeight, float64) {
	var weightSum uint32
	keys := subConns.Keys()
	for _, a := range keys {
		weightSum += getWeightAttribute(a)
	}
	ret := make([]subConnWithWeight, 0, len(keys))
	min := float64(1.0)
	for _, a := range keys {
		v, _ := subConns.Get(a)
		scInfo := v.(*subConn)
		// getWeightAttribute() returns 1 if the weight attribute is not found
		// on the address. And since this function is guaranteed to be called
		// with a non-empty subConns map, weightSum is guaranteed to be
		// non-zero. So, we need not worry about divide a by zero error here.
		nw := float64(getWeightAttribute(a)) / float64(weightSum)
		ret = append(ret, subConnWithWeight{sc: scInfo, weight: nw})
		if nw < min {
			min = nw
		}
	}
	// Sort the addresses to return consistent results.
	//
	// Note: this might not be necessary, but this makes sure the ring is
	// consistent as long as the addresses are the same, for example, in cases
	// where an address is added and then removed, the RPCs will still pick the
	// same old SubConn.
	sort.Slice(ret, func(i, j int) bool { return ret[i].sc.addr < ret[j].sc.addr })
	return ret, min
}

// pick does a binary search. It returns the item with smallest index i that
// r.items[i].hash >= h.
func (r *ring) pick(h uint64) *ringEntry {
	i := sort.Search(len(r.items), func(i int) bool { return r.items[i].hash >= h })
	if i == len(r.items) {
		// If not found, and h is greater than the largest hash, return the
		// first item.
		i = 0
	}
	return r.items[i]
}

// next returns the next entry.
func (r *ring) next(e *ringEntry) *ringEntry {
	return r.items[(e.idx+1)%len(r.items)]
}
