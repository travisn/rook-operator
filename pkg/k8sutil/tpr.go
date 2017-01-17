package k8sutil

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

var initRetryWaitTime = 30 * time.Second

func InitTPR(kclient *unversioned.Client, name, masterHost string) error {
	for {
		err := createTPR(kclient, name, masterHost)
		if err == nil || IsKubernetesResourceAlreadyExistError(err) {
			logger.Infof("TPR %s already exists", name)
			return nil
		}

		logger.Errorf("fail to create TPR %s: %v", name, err)

		logger.Infof("retry tpr in %v...", initRetryWaitTime)
		<-time.After(initRetryWaitTime)
	}

	return nil
}

func createTPR(kclient *unversioned.Client, name, masterHost string) error {
	tpr := &extensions.ThirdPartyResource{
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Versions: []extensions.APIVersion{
			{Name: "v1"},
		},
		Description: "Manage storage clusters",
	}
	_, err := kclient.ThirdPartyResources().Create(tpr)
	if err != nil {
		return err
	}

	return waitRookTPRReady(kclient, 3*time.Second, 30*time.Second, masterHost, Namespace)
}

func waitRookTPRReady(k8s *unversioned.Client, interval, timeout time.Duration, host, ns string) error {
	return Retry(interval, int(timeout/interval), func() (bool, error) {
		resp, err := httpGetClusterSettings(k8s.Client, host, ns)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			return true, nil
		case http.StatusNotFound: // not set up yet. wait.
			return false, nil
		default:
			return false, fmt.Errorf("invalid status code: %v", resp.Status)
		}
	})
}

func httpGetClusterSettings(httpcli *http.Client, host, ns string) (*http.Response, error) {
	if host == "" {
		host = "http://127.0.0.1:8080"
	}

	uri := fmt.Sprintf("%s/apis/coreos.com/v1/namespaces/%s/rook", host, ns)
	return httpcli.Get(uri)
}
