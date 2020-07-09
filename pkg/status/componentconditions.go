package status

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
)

func NewComponentReadyCondition(reason string) *toolchainv1alpha1.Condition {
	currentTime := metav1.Now()
	return &toolchainv1alpha1.Condition{
		Type:               toolchainv1alpha1.ConditionReady,
		Status:             corev1.ConditionTrue,
		Reason:             reason,
		LastTransitionTime: currentTime,
		LastUpdatedTime:    &currentTime,
	}
}

func NewComponentErrorCondition(reason, msg string) *toolchainv1alpha1.Condition {
	currentTime := metav1.Now()
	return &toolchainv1alpha1.Condition{
		Type:               toolchainv1alpha1.ConditionReady,
		Status:             corev1.ConditionFalse,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: currentTime,
		LastUpdatedTime:    &currentTime,
	}
}
