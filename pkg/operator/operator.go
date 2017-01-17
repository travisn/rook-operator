/*
Copyright 2016 The Rook Authors. All rights reserved.

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
package operator

import (
	"fmt"
	"sync"

	k8s "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

//	"k8s.io/kubernetes/pkg/client/unversioned"

type Operator struct {
	Namespace  string
	MasterHost string
	//kclient    *unversioned.Client
	client      *k8s.CoreV1Client
	waitCluster sync.WaitGroup
}

//, kclient *unversioned.Client
func New(namespace string, client *k8s.CoreV1Client) *Operator {
	return &Operator{
		Namespace: namespace,
		//kclient:   kclient,
		client: client,
	}
}

func (o *Operator) Run() error {
	logger.Infof("Creating the namespace %s", o.Namespace)

	// Create the namespace
	ns := &v1.Namespace{}
	ns.Name = o.Namespace
	_, err := o.client.Namespaces().Create(ns)
	if err != nil {
		return fmt.Errorf("failed to create namespace %s. %+v", o.Namespace, err)
	}

	// Start the mon pods
	/*m := mon.New(o.Namespace)
	err := m.Start(o.kclient)
	if err != nil {
		return fmt.Errorf("failed to start the mons. %+v", err)
	}*/

	// Start the OSDs

	return nil
}
