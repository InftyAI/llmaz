/*
Copyright 2024 The InftyAI Team.

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

package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	llmazcoreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	llmazcorev1alpha1 "github.com/inftyai/llmaz/api/core/v1alpha1"
)

var (
	activatorControllerLog = ctrl.Log.WithName("activator-controller")
)

const (
	playgroundsResource     = "playgrounds"
	activatorControllerName = "activator-controller"
)

func tunnel(a, b net.Conn) {
	go io.Copy(a, b)
	go io.Copy(b, a)
}

type Listener interface {
	net.Listener
	Port() int
}

type listener struct {
	net.Listener
	port int
}

func NewListener() (Listener, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	return &listener{
		Listener: l,
		port:     l.Addr().(*net.TCPAddr).Port,
	}, nil
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *listener) Port() int {
	return l.port
}

type Target struct {
	Name      string
	Namespace string
	Port      int
}

type PortInformation struct {
	Target      Target
	Listener    Listener
	Connections []net.Conn
}

type PortManager struct {
	portMap        map[int]*PortInformation
	reversePortMap map[Target]int
	mut            sync.Mutex

	cb func(*PortInformation)
}

func NewPortManager(cb func(*PortInformation)) *PortManager {
	return &PortManager{
		portMap:        map[int]*PortInformation{},
		reversePortMap: map[Target]int{},
		cb:             cb,
	}
}

func (pm *PortManager) AddTarget(name string, namespace string, port int) (*PortInformation, error) {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	target := Target{
		Name:      name,
		Namespace: namespace,
		Port:      port,
	}

	port, ok := pm.reversePortMap[target]
	if ok {
		return pm.portMap[port], nil
	}

	listener, err := NewListener()
	if err != nil {
		return nil, err
	}
	port = listener.Port()
	downstream := &PortInformation{
		Target:   target,
		Listener: listener,
	}
	pm.portMap[port] = downstream
	pm.reversePortMap[target] = port

	go pm.startListener(downstream)
	return downstream, nil
}

func (pm *PortManager) RemoveTarget(name string, namespace string, port int) *PortInformation {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	target := Target{
		Name:      name,
		Namespace: namespace,
		Port:      port,
	}

	port, ok := pm.reversePortMap[target]
	if !ok {
		return nil
	}
	downstream := pm.portMap[port]
	delete(pm.portMap, port)
	delete(pm.reversePortMap, target)
	return downstream
}

func (pm *PortManager) RemoveTargetForAllPorts(name string, namespace string) []*PortInformation {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	var downstreams []*PortInformation
	for port, downstream := range pm.portMap {
		if downstream.Target.Name == name && downstream.Target.Namespace == namespace {
			delete(pm.portMap, port)
			delete(pm.reversePortMap, downstream.Target)
			downstreams = append(downstreams, downstream)
		}
	}
	return downstreams
}

func (pm *PortManager) startListener(downstream *PortInformation) {
	start := false
	for {
		conn, err := downstream.Listener.Accept()
		if err != nil {
			return
		}
		downstream.Connections = append(downstream.Connections, conn)
		if !start {
			go pm.cb(downstream)
			start = true
		}
	}
}
