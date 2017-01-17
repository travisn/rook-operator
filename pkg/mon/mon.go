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
package mon

import (
	"fmt"

	"github.com/rook/rook-operator/pkg/k8sutil"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta/metatypes"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"
)

const (
	dataDir     = "/var/lib/rook/data"
	MonPort     = 6790
	monApp      = "cephmon"
	appAttr     = "app"
	clusterAttr = "rook_cluster"
	monNodeAttr = "mon_node"
	versionAttr = "mon_version"
	tprName     = "mon.rook.io"
)

type Cluster struct {
	Namespace    string
	Keyring      string
	ClusterName  string
	DataDir      string
	Version      string
	MonConfig    []*MonConfig
	MasterHost   string
	Size         int
	Paused       bool
	NodeSelector map[string]string
	AntiAffinity bool
	Port         int32
}

func New(namespace string) *Cluster {
	return &Cluster{
		Namespace: namespace,
		ClusterName: defaultClusterName,
		DataDir:     dataDir,
		Version:     "dev-2017-01-06-e",
		Size:        3,
		Port:        MonPort,
	}
}

func (c *Cluster) Start(kclient *unversioned.Client) error {

	err := k8sutil.InitTPR(kclient, tprName, c.MasterHost)
	if err != nil {
		return err
	}

	logger.Infof("start running one mon")

	err = c.createServiceAndPod(kclient)
	if err != nil {
		return fmt.Errorf("failed to create mon service and pod. %+v", err)
	}

	logger.Infof("started %d/%d mons", 1, c.Size)
	return nil
}

func (c *Cluster) create(kclient *unversioned.Client) error {
	logger.Infof("creating mon cluster: %+v", c)

	if err := c.createClientServiceLB(kclient); err != nil {
		return fmt.Errorf("cluster create: fail to create client service LB: %v", err)
	}
	return nil
}

func (c *Cluster) createClientServiceLB(kclient *unversioned.Client) error {
	if _, err := c.createServiceWithOwner(kclient, c.AsOwner()); err != nil {
		if !k8sutil.IsKubernetesResourceAlreadyExistError(err) {
			return err
		}
	}
	return nil
}

func (c *Cluster) createServiceAndPod(kclient *unversioned.Client) error {

	svc := c.makeService()

	if _, err := c.createService(kclient, svc); err != nil {
		if !k8sutil.IsKubernetesResourceAlreadyExistError(err) {
			return fmt.Errorf("failed to create mon service. %+v", err)
		}
	}

	mon := &MonConfig{Name: "mon0", Port: MonPort}
	pod := c.makeMonPod(mon, c.AsOwner())
	_, err := kclient.Pods(c.Namespace).Create(pod)
	if err != nil {
		return fmt.Errorf("failed to create mon pod for %+v. pod=%+v. %+v", mon, pod, err)
	}
	return nil
}

func (c *Cluster) createService(kclient *unversioned.Client, svc *api.Service) (*api.Service, error) {
	retSvc, err := kclient.Services(c.Namespace).Create(svc)
	if err != nil {
		return nil, err
	}
	return retSvc, nil
}

func (c *Cluster) createServiceWithOwner(kclient *unversioned.Client, owner metatypes.OwnerReference) (*api.Service, error) {
	svc := c.makeService()
	k8sutil.AddOwnerRefToObject(svc.GetObjectMeta(), owner)
	retSvc, err := kclient.Services(c.Namespace).Create(svc)
	if err != nil {
		return nil, err
	}
	return retSvc, nil
}

func (c *Cluster) makeService() *api.Service {
	labels := map[string]string{
		appAttr:     monApp,
		clusterAttr: c.ClusterName,
	}
	svc := &api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   c.ClusterName,
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{
					Name:       "client",
					Port:       c.Port,
					TargetPort: intstr.FromInt(int(c.Port)),
					Protocol:   api.ProtocolTCP,
				},
			},
			Selector: labels,
		},
	}
	return svc
}

func (c *Cluster) makeMonService(mon *MonConfig, owner metatypes.OwnerReference) *api.Service {
	labels := map[string]string{
		appAttr:     monApp,
		monNodeAttr: mon.Name,
		clusterAttr: c.ClusterName,
	}
	svc := &api.Service{
		ObjectMeta: api.ObjectMeta{
			Name:   mon.Name,
			Labels: labels,
		},
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{
					Name:       "client",
					Port:       mon.Port,
					TargetPort: intstr.FromInt(int(mon.Port)),
					Protocol:   api.ProtocolTCP,
				},
			},
			Selector: labels,
		},
	}
	k8sutil.AddOwnerRefToObject(svc.GetObjectMeta(), owner)
	return svc
}

func (c *Cluster) AsOwner() metatypes.OwnerReference {
	//trueVar := true
	// TODO: In 1.5 this is gonna be "k8s.io/kubernetes/pkg/apis/meta/v1"
	// Both api.OwnerReference and metatypes.OwnerReference are combined into that.
	return metatypes.OwnerReference{}
	/*	APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &trueVar,
	}*/
}
