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

package nodesignature

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type workUnit struct {
	Namespace string
	Name      string
}

var stressPods []workUnit

var pods []workUnit
var podsErr error

const (
	clusterMaxNodes       = 5000
	clusterMaxPodsPerNode = 300
)

const (
	stressNamespaceLen = 52
	stressNameLen      = 72
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	var data []byte
	data, podsErr = os.ReadFile(filepath.Join("testdata", "pods.json"))
	if podsErr != nil {
		return
	}
	podsErr = json.Unmarshal(data, &pods)

	stressPodsCount := clusterMaxNodes * clusterMaxPodsPerNode
	for idx := 0; idx < stressPodsCount; idx++ {
		stressPods = append(stressPods, workUnit{
			Namespace: RandStringBytes(stressNamespaceLen),
			Name:      RandStringBytes(stressNameLen),
		})
	}
}

func TestSum(t *testing.T) {
	if len(pods) == 0 || podsErr != nil {
		t.Fatalf("cannot load the test data: %v", podsErr)
	}

	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.Add(pod.Namespace, pod.Name)
	}
	x := ns.Sum()
	if len(x) == 0 {
		t.Fatalf("zero-lenght sum")
	}
}

func TestSumStable(t *testing.T) {
	if len(pods) == 0 || podsErr != nil {
		t.Fatalf("cannot load the test data: %v", podsErr)
	}

	localPods := make([]workUnit, len(pods))
	copy(localPods, pods)
	rand.Shuffle(len(localPods), func(i, j int) {
		localPods[i], localPods[j] = localPods[j], localPods[i]
	})

	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.Add(pod.Namespace, pod.Name)
	}
	nsLocal := &NodeSignature{}
	for _, localPod := range localPods {
		nsLocal.Add(localPod.Namespace, localPod.Name)
	}

	x := ns.Sum()
	xLocal := nsLocal.Sum()
	if !reflect.DeepEqual(x, xLocal) {
		t.Fatalf("signature not stable: %x vs %x", x, xLocal)
	}
}

func TestVerifySign(t *testing.T) {
	if len(pods) == 0 || podsErr != nil {
		t.Fatalf("cannot load the test data: %v", podsErr)
	}

	localPods := make([]workUnit, len(pods))
	copy(localPods, pods)
	rand.Shuffle(len(localPods), func(i, j int) {
		localPods[i], localPods[j] = localPods[j], localPods[i]
	})

	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.Add(pod.Namespace, pod.Name)
	}
	nsLocal := &NodeSignature{}
	for _, localPod := range localPods {
		nsLocal.Add(localPod.Namespace, localPod.Name)
	}

	x := ns.Sign()
	xLocal := nsLocal.Sign()
	if x != xLocal {
		t.Fatalf("signature not stable: %q vs %q", x, xLocal)
	}
}

func stressBenchCluster(maxNodes, maxPodsPerNode int) {
	var nss []*NodeSignature
	for nIdx := 0; nIdx < maxNodes; nIdx++ {
		ns := NewNodeSignature()
		nss = append(nss, ns)
		for pIdx := 0; pIdx < maxPodsPerNode; pIdx++ {
			stressPod := &stressPods[(maxPodsPerNode*nIdx)+pIdx]
			ns.Add(stressPod.Namespace, stressPod.Name)
		}
		_ = ns.Sum()
	}
}

func BenchmarkSignatureStressNode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchCluster(1, clusterMaxPodsPerNode)
	}
}

func BenchmarkSignatureStressCluster1k(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchCluster(1000, clusterMaxPodsPerNode)
	}
}

func BenchmarkSignatureStressFullCluster(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchCluster(clusterMaxNodes, clusterMaxPodsPerNode)
	}
}
