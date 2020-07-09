package status

import (
	"context"
	"fmt"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"

	errs "github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ErrMsgCannotGetDeployment deployment not found
	ErrMsgCannotGetDeployment = "unable to get the operator deployment"

	// ErrMsgDeploymentConditionNotReady deployment not ready
	ErrMsgDeploymentConditionNotReady = "operator deployment has unready status conditions"
)

// GetDeploymentStatusConditions looks up a deployment with the given name within the given namespace and checks its status
// and finally returns a condition summarizing the status and an error if one occurred
func GetDeploymentStatusConditions(client client.Client, name, namespace string) ([]toolchainv1alpha1.Condition, error) {
	deploymentName := types.NamespacedName{Namespace: namespace, Name: name}
	deployment := &appsv1.Deployment{}
	err := client.Get(context.TODO(), deploymentName, deployment)
	if err != nil {
		err = errs.Wrap(err, ErrMsgCannotGetDeployment)
		errCondition := NewComponentErrorCondition(toolchainv1alpha1.ToolchainStatusReasonDeploymentNotFound, err.Error())
		return []toolchainv1alpha1.Condition{*errCondition}, err
	}

	// get and check conditions
	for _, condition := range deployment.Status.Conditions {
		if (condition.Type == appsv1.DeploymentAvailable || condition.Type == appsv1.DeploymentProgressing) && condition.Status != corev1.ConditionTrue {
			// there is a condition that is not ready, return it along with the error
			errMsg := fmt.Sprintf("%s: %s", ErrMsgDeploymentConditionNotReady, condition.Type)
			errCondition := NewComponentErrorCondition(toolchainv1alpha1.ToolchainStatusReasonDeploymentNotReady, errMsg)
			return []toolchainv1alpha1.Condition{*errCondition}, fmt.Errorf(errMsg)
		}
	}

	// no problems with the deployment, return a ready condition
	operatorReadyCondition := NewComponentReadyCondition(toolchainv1alpha1.ToolchainStatusReasonDeploymentReady)
	return []toolchainv1alpha1.Condition{*operatorReadyCondition}, nil
}

func DeploymentReadyCondition() appsv1.DeploymentCondition {
	return appsv1.DeploymentCondition{
		Type:   appsv1.DeploymentAvailable,
		Status: corev1.ConditionTrue,
	}
}

func DeploymentNotReadyCondition() appsv1.DeploymentCondition {
	return appsv1.DeploymentCondition{
		Type:   appsv1.DeploymentAvailable,
		Status: corev1.ConditionFalse,
	}
}

func DeploymentProgressingCondition() appsv1.DeploymentCondition {
	return appsv1.DeploymentCondition{
		Type:   appsv1.DeploymentProgressing,
		Status: corev1.ConditionTrue,
	}
}

func DeploymentNotProgressingCondition() appsv1.DeploymentCondition {
	return appsv1.DeploymentCondition{
		Type:   appsv1.DeploymentProgressing,
		Status: corev1.ConditionFalse,
	}
}
