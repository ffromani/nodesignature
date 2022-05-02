/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodesignature

import (
	"encoding/hex"
	"sort"

	"github.com/OneOfOne/xxhash"
)

const (
	Prefix = "ns"

	Version = "v1"

	Separator = "://"
)

const (
	expectedMaxPodsPerNode = 256
)

type PodIdentificator interface {
	GetNamespace() string
	GetName() string
}

type NodeSignature struct {
	podHashes []uint64
}

func NewNodeSignature() *NodeSignature {
	return &NodeSignature{
		podHashes: make([]uint64, 0, expectedMaxPodsPerNode),
	}
}

func (ns *NodeSignature) Len() int {
	return len(ns.podHashes)
}

func (ns *NodeSignature) AddPod(pi PodIdentificator) error {
	ns.podHashes = append(ns.podHashes,
		xxhash.ChecksumString64S(
			pi.GetName(),
			xxhash.ChecksumString64(pi.GetNamespace()),
		),
	)
	return nil
}

func (ns *NodeSignature) Sum() []byte {
	sort.Sort(uvec64(ns.podHashes))
	h := xxhash.New64()
	b := make([]byte, 8)
	for _, podHash := range ns.podHashes {
		h.Write(putUint64(b, podHash))
	}
	return h.Sum(nil)
}

func (ns *NodeSignature) Sign() string {
	return Prefix + Version + Separator + hex.EncodeToString(ns.Sum())
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
