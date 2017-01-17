package k8sutil

import (
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/meta/metatypes"
)

func AddOwnerRefToObject(o meta.Object, r metatypes.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}
