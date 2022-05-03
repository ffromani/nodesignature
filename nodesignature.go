/*
 * Copyright 2022 Red Hat, Inc.
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

// Package nodesignature computes the signature of a working set.
// A "working set" is a unordered set of namespaced work units
// running at any given time on a "node". The "signature" is a
// unique identifer which is bound to that specific working set.
// This allows to uniquely and concisely identify a workingset
// without enumerating - and storing - all the names all the time.
// This concepts maps nicely on kubernetes (but is not limited to):
// "namespaced work units" = pods and "node" = (kubernetes) nodes.
package nodesignature

import (
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/OneOfOne/xxhash"
)

const (
	// Prefix is the string common to all the signatures
	// A prefix is always 4 bytes long
	Prefix = "nsgn"
	// Version is the version of this signature. You should
	// only compare compatible versions.
	// A Version is always 4 bytes long, in the form v\d\d\d
	Version = "v001"
)

const (
	expectedMaxUnitsPerNode = 256

	expectedPrefixLen  = 4
	expectedVersionLen = 4
	minimumSumLen      = 8
)

var (
	ErrMalformed           = fmt.Errorf("malformed signature")
	ErrIncompatibleVersion = fmt.Errorf("incompatible version")
)

func IsVersionCompatible(ver string) (bool, error) {
	if len(ver) != expectedVersionLen {
		return false, ErrMalformed
	}
	return ver == Version, nil
}

type NodeSignature struct {
	hashes []uint64
}

func NewNodeSignature() *NodeSignature {
	return &NodeSignature{
		hashes: make([]uint64, 0, expectedMaxUnitsPerNode),
	}
}

func (ns *NodeSignature) Len() int {
	return len(ns.hashes)
}

func (ns *NodeSignature) Add(namespace, name string) error {
	ns.hashes = append(ns.hashes,
		xxhash.ChecksumString64S(
			name,
			xxhash.ChecksumString64(namespace),
		),
	)
	return nil
}

func (ns *NodeSignature) Sum() []byte {
	sort.Sort(uvec64(ns.hashes))
	h := xxhash.New64()
	b := make([]byte, 8)
	for _, hash := range ns.hashes {
		h.Write(putUint64(b, hash))
	}
	return h.Sum(nil)
}

func (ns *NodeSignature) Sign() string {
	return Prefix + Version + hex.EncodeToString(ns.Sum())
}

func (ns *NodeSignature) Check(sign string) error {
	if len(sign) < expectedPrefixLen+expectedVersionLen+minimumSumLen {
		return ErrMalformed
	}
	pfx := sign[0:4]
	if pfx != Prefix {
		return ErrMalformed
	}
	ver := sign[4:8]
	ok, err := IsVersionCompatible(ver)
	if err != nil {
		return err
	}
	if !ok {
		return ErrIncompatibleVersion
	}
	sum := sign[8:]
	got := hex.EncodeToString(ns.Sum())
	if got != sum {
		return fmt.Errorf("signature mismatch got=%q expected=%q", got, sum)
	}
	return nil
}

type uvec64 []uint64

func (a uvec64) Len() int           { return len(a) }
func (a uvec64) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a uvec64) Less(i, j int) bool { return a[i] < a[j] }

func putUint64(b []byte, v uint64) []byte {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return b
}
