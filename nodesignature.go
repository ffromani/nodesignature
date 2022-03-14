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
	"crypto/md5"
	"io"
	"sort"
)

type PodIdentificator interface {
	GetNamespace() string
	GetName() string
}

type NodeSignature struct {
	podIdents []string
}

func (ns *NodeSignature) Len() int {
	return len(ns.podIdents)
}

func (ns *NodeSignature) AddPod(pi PodIdentificator) error {
	ns.podIdents = append(ns.podIdents, pi.GetNamespace()+"/"+pi.GetName())
	return nil
}

func (ns *NodeSignature) Sum() []byte {
	sort.Strings(ns.podIdents)
	h := md5.New()
	for _, podIdent := range ns.podIdents {
		io.WriteString(h, podIdent)
	}
	return h.Sum(nil)
}
