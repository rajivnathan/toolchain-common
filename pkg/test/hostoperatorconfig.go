package test

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: delete this file after HostOperatorConfig is removed, see https://issues.redhat.com/browse/CRT-1120
func NewHostOperatorConfig(options ...HostConfigOption) *toolchainv1alpha1.HostOperatorConfig {
	hostOperatorConfig := &toolchainv1alpha1.HostOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: HostOperatorNs,
			Name:      "config",
		},
	}
	for _, option := range options {
		option.Apply(&hostOperatorConfig.Spec)
	}
	return hostOperatorConfig
}
