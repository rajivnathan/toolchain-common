package status

import (
	"testing"
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/test"

	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDeploymentStatusConditions(t *testing.T) {

	t.Run("test deployment status conditions", func(t *testing.T) {
		t.Run("deployment does not exist", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t)
			conditions := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Len(t, conditions, 1)
			require.Equal(t, toolchainv1alpha1.ConditionReady, conditions[0].Type)
			require.Equal(t, corev1.ConditionFalse, conditions[0].Status)
			require.Equal(t, "DeploymentNotFound", conditions[0].Reason)
			require.Contains(t, conditions[0].Message, "unable to get the deployment")
			require.True(t, isRecent(conditions[0].LastTransitionTime), "LastTransitionTime should be recently updated")
			require.True(t, isRecent(*conditions[0].LastUpdatedTime), "LastUpdatedTime should be recently updated")
		})

		t.Run("deployment not available", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentNotAvailable())
			conditions := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Len(t, conditions, 1)
			require.Equal(t, toolchainv1alpha1.ConditionReady, conditions[0].Type)
			require.Equal(t, corev1.ConditionFalse, conditions[0].Status)
			require.Equal(t, "DeploymentNotReady", conditions[0].Reason)
			require.Contains(t, conditions[0].Message, "deployment has unready status conditions")
			require.True(t, isRecent(conditions[0].LastTransitionTime), "LastTransitionTime should be recently updated")
			require.True(t, isRecent(*conditions[0].LastUpdatedTime), "LastUpdatedTime should be recently updated")
		})

		t.Run("deployment not progressing", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentNotProgressing())
			conditions := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Len(t, conditions, 1)
			require.Equal(t, toolchainv1alpha1.ConditionReady, conditions[0].Type)
			require.Equal(t, corev1.ConditionFalse, conditions[0].Status)
			require.Equal(t, "DeploymentNotReady", conditions[0].Reason)
			require.Contains(t, conditions[0].Message, "deployment has unready status conditions")
			require.True(t, isRecent(conditions[0].LastTransitionTime), "LastTransitionTime should be recently updated")
			require.True(t, isRecent(*conditions[0].LastUpdatedTime), "LastUpdatedTime should be recently updated")
		})

		t.Run("deployment ready", func(t *testing.T) {
			fakeClient := test.NewFakeClient(t, fakeDeploymentReady())
			conditions := GetDeploymentStatusConditions(fakeClient, "test-deployment", test.HostOperatorNs)
			require.Len(t, conditions, 1)
			require.Equal(t, toolchainv1alpha1.ConditionReady, conditions[0].Type)
			require.Equal(t, corev1.ConditionTrue, conditions[0].Status)
			require.Equal(t, "DeploymentReady", conditions[0].Reason)
			require.Equal(t, conditions[0].Message, "")
			require.True(t, isRecent(conditions[0].LastTransitionTime), "LastTransitionTime should be recently updated")
			require.True(t, isRecent(*conditions[0].LastUpdatedTime), "LastUpdatedTime should be recently updated")
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

func isRecent(timestamp metav1.Time) bool {
	tenSecondsAgo := metav1.Now().Add(time.Duration(-10) * time.Second)
	return timestamp.After(tenSecondsAgo)
}
