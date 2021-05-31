package test

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HostConfigOptionFunc func(config *toolchainv1alpha1.HostConfig)

type ToolchainConfigOption interface {
	Apply(config *toolchainv1alpha1.ToolchainConfig)
}

type HostConfigOptionImpl struct {
	toApply []HostConfigOptionFunc
}

func (option *HostConfigOptionImpl) Apply(config *toolchainv1alpha1.HostConfig) {
	for _, apply := range option.toApply {
		apply(config)
	}
}

func (option *HostConfigOptionImpl) addFunction(funcToAdd HostConfigOptionFunc) {
	option.toApply = append(option.toApply, funcToAdd)
}

type AutomaticApprovalOption struct {
	*HostConfigOptionImpl
}

func AutomaticApproval() *AutomaticApprovalOption {
	o := &AutomaticApprovalOption{
		HostConfigOptionImpl: &HostConfigOptionImpl{},
	}
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		config.AutomaticApproval = toolchainv1alpha1.AutomaticApproval{}
	})
	return o
}

func (o AutomaticApprovalOption) Enabled() AutomaticApprovalOption {
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		val := true
		config.AutomaticApproval.Enabled = &val
	})
	return o
}

func (o AutomaticApprovalOption) Disabled() AutomaticApprovalOption {
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		val := false
		config.AutomaticApproval.Enabled = &val
	})
	return o
}

type DeactivationOption struct {
	*HostConfigOptionImpl
}

func Deactivation() *DeactivationOption {
	o := &DeactivationOption{
		HostConfigOptionImpl: &HostConfigOptionImpl{},
	}
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		config.Deactivation = toolchainv1alpha1.Deactivation{}
	})
	return o
}

func (o DeactivationOption) DeactivatingNotificationDays(days int) DeactivationOption {
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		config.Deactivation.DeactivatingNotificationDays = &days
	})
	return o
}

type PerMemberClusterOption func(map[string]int)

func PerMemberCluster(name string, value int) PerMemberClusterOption {
	return func(clusters map[string]int) {
		clusters[name] = value
	}
}

func (o AutomaticApprovalOption) ResourceCapThreshold(defaultThreshold int, perMember ...PerMemberClusterOption) AutomaticApprovalOption {
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		config.AutomaticApproval.ResourceCapacityThreshold.DefaultThreshold = &defaultThreshold
		config.AutomaticApproval.ResourceCapacityThreshold.SpecificPerMemberCluster = map[string]int{}
		for _, add := range perMember {
			add(config.AutomaticApproval.ResourceCapacityThreshold.SpecificPerMemberCluster)
		}
	})
	return o
}

func (o AutomaticApprovalOption) MaxUsersNumber(overall int, perMember ...PerMemberClusterOption) AutomaticApprovalOption {
	o.addFunction(func(config *toolchainv1alpha1.HostConfig) {
		config.AutomaticApproval.MaxNumberOfUsers.Overall = &overall
		config.AutomaticApproval.MaxNumberOfUsers.SpecificPerMemberCluster = map[string]int{}
		for _, add := range perMember {
			add(config.AutomaticApproval.MaxNumberOfUsers.SpecificPerMemberCluster)
		}
	})
	return o
}

func NewToolchainConfig(options ...ToolchainConfigOption) *toolchainv1alpha1.ToolchainConfig {
	toolchainConfig := &toolchainv1alpha1.ToolchainConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: HostOperatorNs,
			Name:      "config",
		},
	}
	for _, option := range options {
		option.Apply(toolchainConfig)
	}
	return toolchainConfig
}
