package k8sutil

import (
	"fmt"
	"net/http"

	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	unversionedCli "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/types"

	apierrors "k8s.io/kubernetes/pkg/api/errors"
)

const (
	Namespace = "rook"
)

func IsKubernetesResourceAlreadyExistError(err error) bool {
	se, ok := err.(*apierrors.StatusError)
	if !ok {
		return false
	}
	if se.Status().Code == http.StatusConflict && se.Status().Reason == unversioned.StatusReasonAlreadyExists {
		return true
	}
	return false
}

func IsKubernetesResourceNotFoundError(err error) bool {
	se, ok := err.(*apierrors.StatusError)
	if !ok {
		return false
	}
	if se.Status().Code == http.StatusNotFound && se.Status().Reason == unversioned.StatusReasonNotFound {
		return true
	}
	return false
}

func PollPods(kubeCli *unversionedCli.Client, clusterName, namespace, app string, uid types.UID) ([]*k8sapi.Pod, []*k8sapi.Pod, error) {
	podList, err := kubeCli.Pods(namespace).List(ClusterListOpt(app, clusterName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list running pods: %v", err)
	}

	var running []*k8sapi.Pod
	var pending []*k8sapi.Pod
	for i := range podList.Items {
		pod := &podList.Items[i]
		if len(pod.OwnerReferences) < 1 {
			logger.Warningf("pollPods: ignore pod %v: no owner", pod.Name)
			continue
		}
		if pod.OwnerReferences[0].UID != uid {
			logger.Warningf("pollPods: ignore pod %v: owner (%v) is not %v", pod.Name, pod.OwnerReferences[0].UID, uid)
			continue
		}
		switch pod.Status.Phase {
		case k8sapi.PodRunning:
			running = append(running, pod)
		case k8sapi.PodPending:
			pending = append(pending, pod)
		}
	}

	return running, pending, nil
}

func ClusterListOpt(app, clusterName string) k8sapi.ListOptions {
	return k8sapi.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"rook_cluster": clusterName,
			"app":          app,
		}),
	}
}

func GetPodNames(pods []*k8sapi.Pod) []string {
	res := []string{}
	for _, p := range pods {
		res = append(res, p.Name)
	}
	return res
}

func MakeRookImage(version string) string {
	return fmt.Sprintf("quay.io/rook/rook:%v", version)
}
