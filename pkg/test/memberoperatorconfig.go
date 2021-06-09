package test

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MemberOperatorConfigOptionFunc func(config *toolchainv1alpha1.MemberOperatorConfig)

type MemberOperatorConfigOption interface {
	Apply(config *toolchainv1alpha1.MemberOperatorConfig)
}

type MemberOperatorConfigOptionImpl struct {
	toApply []MemberOperatorConfigOptionFunc
}

func (option *MemberOperatorConfigOptionImpl) Apply(config *toolchainv1alpha1.MemberOperatorConfig) {
	for _, apply := range option.toApply {
		apply(config)
	}
}

func (option *MemberOperatorConfigOptionImpl) addFunction(funcToAdd MemberOperatorConfigOptionFunc) {
	option.toApply = append(option.toApply, funcToAdd)
}

type MemberStatusOption struct {
	*MemberOperatorConfigOptionImpl
}

func MemberStatus() *MemberStatusOption {
	o := &MemberStatusOption{
		MemberOperatorConfigOptionImpl: &MemberOperatorConfigOptionImpl{},
	}
	o.addFunction(func(config *toolchainv1alpha1.MemberOperatorConfig) {
		config.Spec.MemberStatus = toolchainv1alpha1.MemberStatusConfig{}
	})
	return o
}

func (o MemberStatusOption) RefreshPeriod(refreshPeriod string) MemberStatusOption {
	o.addFunction(func(config *toolchainv1alpha1.MemberOperatorConfig) {
		config.Spec.MemberStatus.RefreshPeriod = &refreshPeriod
	})
	return o
}

func NewMemberOperatorConfig(options ...MemberOperatorConfigOption) *toolchainv1alpha1.MemberOperatorConfig {
	memberOperatorConfig := &toolchainv1alpha1.MemberOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: MemberOperatorNs,
			Name:      "config",
		},
	}
	for _, option := range options {
		option.Apply(memberOperatorConfig)
	}
	return memberOperatorConfig
}
