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
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type podIdent struct {
	Namespace string
	Name      string
}

var stressPods []podIdent

var pods []podIdent
var podsErr error

const (
	stressPodsCount    = 8192
	stressNamespaceLen = 24
	stressNameLen      = 48
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

	for idx := 0; idx < stressPodsCount; idx++ {
		stressPods = append(stressPods, podIdent{
			Namespace: RandStringBytes(stressNamespaceLen),
			Name:      RandStringBytes(stressNameLen),
		})
	}
}

func TestSignature(t *testing.T) {
	if len(pods) == 0 || podsErr != nil {
		t.Fatalf("cannot load the test data: %v", podsErr)
	}

	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.AddPod(pod.Namespace, pod.Name)
	}
	x := ns.Sum()
	if len(x) == 0 {
		t.Fatalf("zero-lenght sum")
	}
}

func TestSignatureStable(t *testing.T) {
	if len(pods) == 0 || podsErr != nil {
		t.Fatalf("cannot load the test data: %v", podsErr)
	}

	localPods := make([]podIdent, len(pods))
	copy(localPods, pods)
	rand.Shuffle(len(localPods), func(i, j int) {
		localPods[i], localPods[j] = localPods[j], localPods[i]
	})

	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.AddPod(pod.Namespace, pod.Name)
	}
	nsLocal := &NodeSignature{}
	for _, localPod := range localPods {
		nsLocal.AddPod(localPod.Namespace, localPod.Name)
	}

	x := ns.Sum()
	xLocal := nsLocal.Sum()
	if !reflect.DeepEqual(x, xLocal) {
		t.Fatalf("signature not stable: %x vs %x", x, xLocal)
	}
}

func benchHelper() {
	ns := &NodeSignature{}
	for _, pod := range pods {
		ns.AddPod(pod.Namespace, pod.Name)
	}
	_ = ns.Sum()
}

func BenchmarkSignature(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchHelper()
	}
}

func stressBenchHelper(count int) {
	ns := &NodeSignature{}
	for idx := 0; idx < count; idx++ {
		stressPod := &stressPods[idx]
		ns.AddPod(stressPod.Namespace, stressPod.Name)
	}
	_ = ns.Sum()
}

func BenchmarkSignatureStress256(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchHelper(256)
	}
}

func BenchmarkSignatureStress512(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchHelper(512)
	}
}

func BenchmarkSignatureStress1024(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchHelper(1024)
	}
}

func BenchmarkSignatureStress2048(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stressBenchHelper(2048)
	}
}
