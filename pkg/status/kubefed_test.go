package status

import (
	"testing"
	"time"

	"github.com/codeready-toolchain/toolchain-common/pkg/cluster"
	"github.com/codeready-toolchain/toolchain-common/pkg/test"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kubefed/pkg/apis/core/common"
	kubefed_common "sigs.k8s.io/kubefed/pkg/apis/core/common"
	kubefed_v1beta1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
)

func TestGetKubefedConditions(t *testing.T) {
	t.Run("test kubefed conditions", func(t *testing.T) {
		t.Run("condition ready", func(t *testing.T) {
			testReason := "ReasonComesFromKubefed"
			readyAttrs := KubefedAttributes{
				GetClusterFunc: newGetHostClusterReady(&testReason),
				Period:         10 * time.Second,
				Timeout:        3 * time.Second,
				Threshold:      3,
			}
			testClusterConditions, err := GetKubefedConditions(readyAttrs)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(testClusterConditions.Conditions))
			for _, testCondition := range testClusterConditions.Conditions {
				assert.Equal(t, testReason, *testCondition.Reason)
				assert.Nil(t, testCondition.Message, "message should be nil")
				assert.Equal(t, corev1.ConditionTrue, testCondition.Status)
				assert.Equal(t, kubefed_common.ClusterReady, testCondition.Type)
			}
		})

		t.Run("condition cluster not ok", func(t *testing.T) {
			expectedReason := "KubefedNotFound"
			msg := "the cluster connection was not found"
			readyAttrs := KubefedAttributes{
				GetClusterFunc: newGetHostClusterNotOk(&expectedReason, &msg),
				Period:         10 * time.Second,
				Timeout:        3 * time.Second,
				Threshold:      3,
			}
			testClusterConditions, err := GetKubefedConditions(readyAttrs)
			assert.Error(t, err)
			assert.Equal(t, msg, err.Error())
			assert.Equal(t, 1, len(testClusterConditions.Conditions))
			for _, testCondition := range testClusterConditions.Conditions {
				assert.Equal(t, expectedReason, *testCondition.Reason)
				assert.Equal(t, msg, *testCondition.Message)
				assert.Equal(t, corev1.ConditionFalse, testCondition.Status)
				assert.Equal(t, kubefed_common.ClusterReady, testCondition.Type)
			}
		})

		t.Run("condition cluster ok but not ready", func(t *testing.T) {
			testReason := "ReasonComesFromKubefed"
			msg := "the cluster connection is not ready"
			readyAttrs := KubefedAttributes{
				GetClusterFunc: newGetHostClusterOkButNotReady(&testReason, &msg),
				Period:         10 * time.Second,
				Timeout:        3 * time.Second,
				Threshold:      3,
			}
			testClusterConditions, err := GetKubefedConditions(readyAttrs)
			assert.Error(t, err)
			assert.Equal(t, msg, err.Error())
			assert.Equal(t, 1, len(testClusterConditions.Conditions))
			for _, testCondition := range testClusterConditions.Conditions {
				assert.Equal(t, testReason, *testCondition.Reason)
				assert.Equal(t, msg, *testCondition.Message)
				assert.Equal(t, corev1.ConditionFalse, testCondition.Status)
				assert.Equal(t, kubefed_common.ClusterReady, testCondition.Type)
			}
		})

		t.Run("condition last probe time exceeded", func(t *testing.T) {
			testReason := "KubefedLastProbeTimeExceeded"
			msg := "exceeded the maximum duration since the last probe"
			readyAttrs := KubefedAttributes{
				GetClusterFunc: newGetHostClusterLastProbeTimeExceeded(&testReason, &msg),
				Period:         10 * time.Second,
				Timeout:        3 * time.Second,
				Threshold:      3,
			}
			testClusterConditions, err := GetKubefedConditions(readyAttrs)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), msg)
			assert.Equal(t, 1, len(testClusterConditions.Conditions))
			for _, testCondition := range testClusterConditions.Conditions {
				assert.Equal(t, testReason, *testCondition.Reason)
				assert.Contains(t, *testCondition.Message, msg)
				assert.Equal(t, corev1.ConditionFalse, testCondition.Status)
				assert.Equal(t, kubefed_common.ClusterReady, testCondition.Type)
			}
		})
	})
}

func newGetHostClusterReady(reason *string) cluster.GetHostClusterFunc {
	msg := ""
	return NewFakeGetHostCluster(true, corev1.ConditionTrue, metav1.Now(), reason, &msg)
}

func newGetHostClusterNotOk(reason, message *string) cluster.GetHostClusterFunc {
	return NewFakeGetHostCluster(false, corev1.ConditionFalse, metav1.Now(), reason, message)
}

func newGetHostClusterOkButNotReady(reason, message *string) cluster.GetHostClusterFunc {
	return NewFakeGetHostCluster(true, corev1.ConditionFalse, metav1.Now(), reason, message)
}

func newGetHostClusterLastProbeTimeExceeded(reason, message *string) cluster.GetHostClusterFunc {
	tenMinsAgo := metav1.Now().Add(time.Duration(-10) * time.Minute)
	return NewFakeGetHostCluster(true, corev1.ConditionTrue, metav1.NewTime(tenMinsAgo), reason, message)
}

// NewGetHostCluster returns cluster.GetHostClusterFunc function. The cluster.FedCluster
// that is returned by the function then contains the given client and the given status and lastProbeTime.
// If ok == false, then the function returns nil for the cluster.
func NewFakeGetHostCluster(ok bool, status corev1.ConditionStatus, lastProbeTime metav1.Time, reason, message *string) cluster.GetHostClusterFunc {
	if !ok {
		return func() (*cluster.FedCluster, bool) {
			return nil, false
		}
	}
	return func() (*cluster.FedCluster, bool) {
		fedClusterValue := &cluster.FedCluster{
			Type:              cluster.Host,
			OperatorNamespace: test.HostOperatorNs,
			OwnerClusterName:  test.MemberClusterName,
			ClusterStatus: &kubefed_v1beta1.KubeFedClusterStatus{
				Conditions: []kubefed_v1beta1.ClusterCondition{{
					Type:          common.ClusterReady,
					Reason:        reason,
					Status:        status,
					LastProbeTime: lastProbeTime,
				}},
			},
		}
		if *message != "" {
			fedClusterValue.ClusterStatus.Conditions[0].Message = message
		}

		return fedClusterValue, true
	}
}
