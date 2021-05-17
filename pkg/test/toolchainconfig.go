package test

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/pkg/apis/toolchain/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ToolchainConfigOptionFunc func(config *toolchainv1alpha1.ToolchainConfig)

type ToolchainConfigOption interface {
	Apply(config *toolchainv1alpha1.ToolchainConfig)
}

type ToolchainConfigOptionImpl struct {
	toApply []ToolchainConfigOptionFunc
}

func (option *ToolchainConfigOptionImpl) Apply(config *toolchainv1alpha1.ToolchainConfig) {
	for _, apply := range option.toApply {
		apply(config)
	}
}

func (option *ToolchainConfigOptionImpl) addFunction(funcToAdd ToolchainConfigOptionFunc) {
	option.toApply = append(option.toApply, funcToAdd)
}

type AutomaticApprovalCfgOption struct {
	*ToolchainConfigOptionImpl
}

func AutomaticApprovalCfg() *AutomaticApprovalCfgOption {
	o := &AutomaticApprovalCfgOption{
		ToolchainConfigOptionImpl: &ToolchainConfigOptionImpl{},
	}
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		config.Spec.Host.AutomaticApproval = toolchainv1alpha1.AutomaticApprovalCfg{}
	})
	return o
}

func (o AutomaticApprovalCfgOption) EnabledCfg() AutomaticApprovalCfgOption {
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		val := true
		config.Spec.Host.AutomaticApproval.Enabled = &val
	})
	return o
}

func (o AutomaticApprovalCfgOption) DisabledCfg() AutomaticApprovalCfgOption {
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		val := false
		config.Spec.Host.AutomaticApproval.Enabled = &val
	})
	return o
}

type DeactivationCfgOption struct {
	*ToolchainConfigOptionImpl
}

func DeactivationCfg() *DeactivationCfgOption {
	o := &DeactivationCfgOption{
		ToolchainConfigOptionImpl: &ToolchainConfigOptionImpl{},
	}
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		config.Spec.Host.Deactivation = toolchainv1alpha1.DeactivationCfg{}
	})
	return o
}

func (o DeactivationCfgOption) DeactivatingNotificationDays(days int) DeactivationCfgOption {
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		config.Spec.Host.Deactivation.DeactivatingNotificationDays = &days
	})
	return o
}

type PerMemberClusterCfgOption func(map[string]int)

func PerMemberClusterCfg(name string, value int) PerMemberClusterCfgOption {
	return func(clusters map[string]int) {
		clusters[name] = value
	}
}

func (o AutomaticApprovalCfgOption) ResourceCapThreshold(defaultThreshold int, perMember ...PerMemberClusterCfgOption) AutomaticApprovalCfgOption {
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		config.Spec.Host.AutomaticApproval.ResourceCapacityThreshold.DefaultThreshold = &defaultThreshold
		config.Spec.Host.AutomaticApproval.ResourceCapacityThreshold.SpecificPerMemberCluster = map[string]int{}
		for _, add := range perMember {
			add(config.Spec.Host.AutomaticApproval.ResourceCapacityThreshold.SpecificPerMemberCluster)
		}
	})
	return o
}

func (o AutomaticApprovalCfgOption) MaxUsersNumber(overall int, perMember ...PerMemberClusterCfgOption) AutomaticApprovalCfgOption {
	o.addFunction(func(config *toolchainv1alpha1.ToolchainConfig) {
		config.Spec.Host.AutomaticApproval.MaxNumberOfUsers.Overall = &overall
		config.Spec.Host.AutomaticApproval.MaxNumberOfUsers.SpecificPerMemberCluster = map[string]int{}
		for _, add := range perMember {
			add(config.Spec.Host.AutomaticApproval.MaxNumberOfUsers.SpecificPerMemberCluster)
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
