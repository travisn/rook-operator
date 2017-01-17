package mon

import (
	"fmt"
	"strings"

	"github.com/rook/rook-operator/pkg/k8sutil"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta/metatypes"
)

func (c *Cluster) makeMonPod(mon *MonConfig, owner metatypes.OwnerReference) *api.Pod {

	container := c.monContainer(mon)
	container.LivenessProbe = mon.livenessProbe()

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: mon.Name,
			Labels: map[string]string{
				appAttr:     monApp,
				monNodeAttr: mon.Name,
				clusterAttr: c.ClusterName,
			},
			Annotations: map[string]string{},
		},
		Spec: api.PodSpec{
			Containers:    []api.Container{container},
			RestartPolicy: api.RestartPolicyNever,
			Volumes: []api.Volume{
				{Name: "rook-data", VolumeSource: api.VolumeSource{EmptyDir: &api.EmptyDirVolumeSource{}}},
			},
		},
	}

	k8sutil.SetPodVersion(pod, versionAttr, c.Version)

	if c.AntiAffinity {
		pod = k8sutil.PodWithAntiAffinity(pod, c.ClusterName)
	}

	if len(c.NodeSelector) != 0 {
		pod.Spec.NodeSelector = c.NodeSelector
	}

	//k8sutil.AddOwnerRefToObject(pod.GetObjectMeta(), owner)
	return pod
}

func (c *Cluster) monContainer(mon *MonConfig) api.Container {
	command := fmt.Sprintf("/usr/local/bin/rook mon --data-dir=%s --name=%s --initial-mons=%s ",
		c.DataDir, mon.Name, strings.Join(mon.InitialMons, ","))

	return api.Container{
		// TODO: fix "sleep 5".
		// Without waiting some time, there is highly probable flakes in network setup.
		Command: []string{"/bin/sh", "-c", fmt.Sprintf("sleep 5; %s", command)},
		Name:    "cephmon",
		Image:   k8sutil.MakeRookImage(c.Version),
		Ports: []api.ContainerPort{
			{
				Name:          "client",
				ContainerPort: mon.Port,
				Protocol:      api.ProtocolTCP,
			},
		},
		VolumeMounts: []api.VolumeMount{
			{Name: "rook-data", MountPath: c.DataDir},
		},
	}
}

func (m *MonConfig) livenessProbe() *api.Probe {
	// simple query of the REST api locally to see if the pod is alive
	return &api.Probe{
		Handler: api.Handler{
			Exec: &api.ExecAction{
				Command: []string{"/bin/sh", "-c", "curl localhost:8124"},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       60,
		FailureThreshold:    3,
	}
}
