package status

import (
	"fmt"
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/cluster"

	errs "github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefed_common "sigs.k8s.io/kubefed/pkg/apis/core/common"
	kubefed_v1beta1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	kubefed_util "sigs.k8s.io/kubefed/pkg/controller/util"
)

// error messages related to cluster connection
const (
	ErrMsgClusterConnectionNotFound              = "the cluster connection was not found"
	ErrMsgClusterConnectionLastProbeTimeExceeded = "exceeded the maximum duration since the last probe"
)

// KubefedAttributes required attributes for obtaining kubefed status
type KubefedAttributes struct {
	GetClusterFunc func() (*cluster.FedCluster, bool)
	Period         time.Duration
	Timeout        time.Duration
	Threshold      int64
}

// GetKubefedConditions uses the provided kubefed attributes to determine kubefed status
func GetKubefedConditions(attrs KubefedAttributes) (*kubefed_v1beta1.KubeFedClusterStatus, error) {
	// look up cluster connection status
	fedCluster, ok := attrs.GetClusterFunc()
	if !ok {
		notFoundMsg := ErrMsgClusterConnectionNotFound
		notFoundReason := toolchainv1alpha1.ToolchainStatusClusterConnectionNotFoundReason
		badClusterStatus := kubefed_v1beta1.KubeFedClusterStatus{
			Conditions: []kubefed_v1beta1.ClusterCondition{
				{
					Type:          kubefed_common.ClusterReady,
					Status:        corev1.ConditionFalse,
					Reason:        &notFoundReason,
					Message:       &notFoundMsg,
					LastProbeTime: metav1.Now(),
				},
			},
		}
		return &badClusterStatus, fmt.Errorf(notFoundMsg)
	}
	clusterStatus := *fedCluster.ClusterStatus.DeepCopy()

	// check conditions of cluster connection
	if !kubefed_util.IsClusterReady(fedCluster.ClusterStatus) {
		return &clusterStatus, fmt.Errorf("the cluster connection is not ready")
	}

	var lastProbeTime metav1.Time
	foundLastProbeTime := false
	for _, condition := range fedCluster.ClusterStatus.Conditions {
		if condition.Type == kubefed_common.ClusterReady {
			lastProbeTime = condition.LastProbeTime
			foundLastProbeTime = true
		}
	}
	if !foundLastProbeTime {
		return &clusterStatus, fmt.Errorf("the time of the last probe could not be determined")
	}

	// check that the last probe time is within limits. It should be less than (period + timeout) * threshold

	totalf := (attrs.Period.Seconds() + attrs.Timeout.Seconds()) * float64(attrs.Threshold)
	maxDuration, err := time.ParseDuration(fmt.Sprintf("%fs", totalf))
	if err != nil {
		return &clusterStatus, errs.Wrap(err, "the maximum duration since the last probe could not be determined - check the configured values for the kubefed health check period, timeout and failure threshold")
	}

	lastProbedTimePlusMaxDuration := lastProbeTime.Add(maxDuration)
	currentTime := time.Now()
	if currentTime.After(lastProbedTimePlusMaxDuration) {
		errMsg := fmt.Sprintf("%s: %s", ErrMsgClusterConnectionLastProbeTimeExceeded, maxDuration.String())
		errReason := toolchainv1alpha1.ToolchainStatusClusterConnectionLastProbeTimeExceededReason
		badProbeCondition := kubefed_v1beta1.KubeFedClusterStatus{
			Conditions: []kubefed_v1beta1.ClusterCondition{
				{
					Type:          kubefed_common.ClusterReady,
					Status:        corev1.ConditionFalse,
					Reason:        &errReason,
					Message:       &errMsg,
					LastProbeTime: lastProbeTime,
				},
			},
		}
		return &badProbeCondition, fmt.Errorf(errMsg)
	}
	return &clusterStatus, nil
}
