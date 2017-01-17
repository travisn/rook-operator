package mon

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
)

type ClusterPhase string

const (
	defaultClusterName = "rookcluster"

	ClusterPhaseNone     ClusterPhase = ""
	ClusterPhaseCreating              = "Creating"
	ClusterPhaseRunning               = "Running"
	ClusterPhaseFailed                = "Failed"

	ClusterConditionReady              = "Ready"
	ClusterConditionRemovingDeadMember = "RemovingDeadMember"
	ClusterConditionRecovering         = "Recovering"
	ClusterConditionScalingUp          = "ScalingUp"
	ClusterConditionScalingDown        = "ScalingDown"
	ClusterConditionUpgrading          = "Upgrading"
)

type MonConfig struct {
	Name        string
	Port        int32
	InitialMons []string
}

func (m *MonConfig) monLivenessProbe() *api.Probe {
	// mon pod is alive only if a query succeeds.
	return &api.Probe{
		Handler: api.Handler{
			Exec: &api.ExecAction{
				Command: []string{"/bin/curl", fmt.Sprintf("localhost:%d", m.Port)},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		PeriodSeconds:       60,
		FailureThreshold:    3,
	}
}
