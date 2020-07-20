package status

import (
	"testing"
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDeploymentStatusConditions(t *testing.T) {

	t.Run("test deployment status conditions", func(t *testing.T) {
		t.Run("deployment does not exist", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t)
			conditions, err := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Error(t, err)

			expected := toolchainv1alpha1.Condition{
				Type:    toolchainv1alpha1.ConditionReady,
				Status:  corev1.ConditionFalse,
				Reason:  "DeploymentNotFound",
				Message: "unable to get the deployment: deployments.apps \"test-deployment\" not found",
			}
			assertConditionsMatchAndRecentTimestamps(t, conditions, expected)
		})

		t.Run("deployment not available", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentNotAvailable())
			conditions, err := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Error(t, err)
			expected := toolchainv1alpha1.Condition{
				Type:    toolchainv1alpha1.ConditionReady,
				Status:  corev1.ConditionFalse,
				Reason:  "DeploymentNotReady",
				Message: "deployment has unready status conditions: Available",
			}
			assertConditionsMatchAndRecentTimestamps(t, conditions, expected)
		})

		t.Run("deployment not progressing", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentNotProgressing())
			conditions, err := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Error(t, err)
			expected := toolchainv1alpha1.Condition{
				Type:    toolchainv1alpha1.ConditionReady,
				Status:  corev1.ConditionFalse,
				Reason:  "DeploymentNotReady",
				Message: "deployment has unready status conditions: Progressing",
			}
			assertConditionsMatchAndRecentTimestamps(t, conditions, expected)
		})

		t.Run("deployment ready", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentReady())
			conditions, err := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.NoError(t, err)
			expected := toolchainv1alpha1.Condition{
				Type:    toolchainv1alpha1.ConditionReady,
				Status:  corev1.ConditionTrue,
				Reason:  "DeploymentReady",
				Message: "",
			}
			assertConditionsMatchAndRecentTimestamps(t, conditions, expected)
		})
	})
}

func fakeDeploymentNotAvailable() *appsv1.Deployment {
	return newFakeDeployment("test-deployment", test.HostOperatorNs, DeploymentNotAvailableCondition(), DeploymentProgressingCondition())
}

func fakeDeploymentNotProgressing() *appsv1.Deployment {
	return newFakeDeployment("test-deployment", test.HostOperatorNs, DeploymentAvailableCondition(), DeploymentNotProgressingCondition())
}

func fakeDeploymentReady() *appsv1.Deployment {
	return newFakeDeployment("test-deployment", test.HostOperatorNs, DeploymentAvailableCondition(), DeploymentProgressingCondition())
}

func newFakeDeployment(name, namespace string, deploymentConditions ...appsv1.DeploymentCondition) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			Conditions: deploymentConditions,
		},
	}
}

func assertConditionsMatchAndRecentTimestamps(t *testing.T, actual []toolchainv1alpha1.Condition, expected ...toolchainv1alpha1.Condition) {
	test.AssertConditionsMatch(t, actual, expected...)
	assertTimestampsAreRecent(t, actual)
}

func assertTimestampsAreRecent(t *testing.T, actual []toolchainv1alpha1.Condition) {
	for _, c := range actual {
		assert.True(t, isRecent(c.LastTransitionTime))
		assert.True(t, isRecent(*c.LastUpdatedTime))
	}
}

func isRecent(timestamp metav1.Time) bool {
	tenSecondsAgo := metav1.Now().Add(time.Duration(-10) * time.Second)
	return timestamp.After(tenSecondsAgo)
}
