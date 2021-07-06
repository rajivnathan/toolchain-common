package configuration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/codeready-toolchain/toolchain-common/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestLoadFromConfigMap(t *testing.T) {
	restore := test.SetEnvVarAndRestore(t, "WATCH_NAMESPACE", "toolchain-member-operator")
	defer restore()

	t.Run("configMap not found", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "MEMBER_OPERATOR_CONFIG_MAP_NAME", "test-config")
		defer restore()

		cl := test.NewFakeClient(t)

		// when
		err := LoadFromConfigMap("MEMBER_OPERATOR", "MEMBER_OPERATOR_CONFIG_MAP_NAME", cl)

		// then
		require.NoError(t, err)
	})
	t.Run("no config name set", func(t *testing.T) {
		// given
		data := map[string]string{
			"super-special-key": "super-special-value",
		}
		cl := test.NewFakeClient(t, createConfigMap("test-config", "toolchain-host-operator", data))

		// when
		err := LoadFromConfigMap("HOST_OPERATOR", "HOST_OPERATOR_CONFIG_MAP_NAME", cl)

		// then
		require.NoError(t, err)

		// test that the secret was not found since no secret name was set
		testTest := os.Getenv("HOST_OPERATOR_SUPER_SPECIAL_KEY")
		assert.Equal(t, "", testTest)
	})
	t.Run("cannot get configmap", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "MEMBER_OPERATOR_CONFIG_MAP_NAME", "test-config")
		defer restore()

		data := map[string]string{
			"test-key-one": "test-value-one",
		}
		cl := test.NewFakeClient(t, createConfigMap("test-config", "toolchain-host-operator", data))

		cl.MockGet = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
			return errors.New("oopsie woopsie")
		}

		// when
		err := LoadFromConfigMap("MEMBER_OPERATOR", "MEMBER_OPERATOR_CONFIG_MAP_NAME", cl)

		// then
		require.Error(t, err)
		assert.Equal(t, "oopsie woopsie", err.Error())

		// test env vars are parsed and created correctly
		testTest := os.Getenv("MEMBER_OPERATOR_TEST_KEY_ONE")
		assert.Equal(t, testTest, "")
	})
	t.Run("env overwrite", func(t *testing.T) {
		// given
		restore := test.SetEnvVarsAndRestore(t,
			test.Env("MEMBER_OPERATOR_CONFIG_MAP_NAME", "test-config"),
			test.Env("MEMBER_OPERATOR_TEST_KEY", ""))
		defer restore()

		data := map[string]string{
			"test-key": "test-value",
		}
		cl := test.NewFakeClient(t, createConfigMap("test-config", "toolchain-member-operator", data))

		// when
		err := LoadFromConfigMap("MEMBER_OPERATOR", "MEMBER_OPERATOR_CONFIG_MAP_NAME", cl)

		// then
		require.NoError(t, err)

		// test env vars are parsed and created correctly
		testTest := os.Getenv("MEMBER_OPERATOR_TEST_KEY")
		assert.Equal(t, testTest, "test-value")
	})
}

func TestLoadFromSecret(t *testing.T) {
	restore := test.SetEnvVarAndRestore(t, "WATCH_NAMESPACE", "toolchain-host-operator")
	defer restore()
	t.Run("secret not found", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "HOST_OPERATOR_SECRET_NAME", "test-secret")
		defer restore()

		cl := test.NewFakeClient(t)

		// when
		secretData, err := LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)

		// then
		require.NoError(t, err)
		assert.Empty(t, secretData)
	})
	t.Run("no secret name set", func(t *testing.T) {
		// given
		data := map[string][]byte{
			"special.key": []byte("special-value"),
		}
		cl := test.NewFakeClient(t, test.CreateSecret("test-secret", "toolchain-host-operator", data))

		// when
		secretData, err := LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)

		// then
		require.NoError(t, err)
		assert.Empty(t, secretData)
	})
	t.Run("cannot get secret", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "HOST_OPERATOR_SECRET_NAME", "test-secret")
		defer restore()

		data := map[string][]byte{
			"test.key.secret": []byte("test-value-secret"),
		}
		cl := test.NewFakeClient(t, test.CreateSecret("test-secret", "toolchain-host-operator", data))

		cl.MockGet = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
			return errors.New("oopsie woopsie")
		}

		// when
		secretData, err := LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)

		// then
		require.Error(t, err)
		assert.Equal(t, "oopsie woopsie", err.Error())
		assert.Empty(t, secretData)
	})
	t.Run("env overwrite", func(t *testing.T) {
		// given
		restore := test.SetEnvVarsAndRestore(t,
			test.Env("HOST_OPERATOR_SECRET_NAME", "test-secret"),
			test.Env("HOST_OPERATOR_TEST_KEY", ""))
		defer restore()

		data := map[string][]byte{
			"test.key": []byte("test-value"),
		}
		cl := test.NewFakeClient(t, test.CreateSecret("test-secret", "toolchain-host-operator", data))

		// when
		secretData, err := LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)

		// then
		require.NoError(t, err)

		// test env vars are parsed and created correctly
		assert.Equal(t, 1, len(secretData))
		assert.Equal(t, "test-value", secretData["test.key"])
	})
}

func TestNoWatchNamespaceSetWhenLoadingSecret(t *testing.T) {
	t.Run("no watch namespace", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "HOST_OPERATOR_SECRET_NAME", "test-secret")
		defer restore()

		data := map[string][]byte{
			"test.key": []byte("test-value"),
		}
		cl := test.NewFakeClient(t, test.CreateSecret("test-secret", "toolchain-host-operator", data))

		// when
		secretData, err := LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)

		// then
		require.Error(t, err)
		assert.Empty(t, secretData)
		assert.Equal(t, "WATCH_NAMESPACE must be set", err.Error())
	})
}

func TestNoWatchNamespaceSetWhenLoadingConfigMap(t *testing.T) {
	t.Run("no watch namespace", func(t *testing.T) {
		// given
		restore := test.SetEnvVarAndRestore(t, "HOST_OPERATOR_CONFIG_MAP_NAME", "test-config")
		defer restore()

		data := map[string]string{
			"test-key": "test-value",
		}
		cl := test.NewFakeClient(t, createConfigMap("test-config", "toolchain-host-operator", data))

		// when
		err := LoadFromConfigMap("HOST_OPERATOR", "HOST_OPERATOR_CONFIG_MAP_NAME", cl)

		// then
		require.Error(t, err)
		assert.Equal(t, "WATCH_NAMESPACE must be set", err.Error())
	})
}

func TestLoadSecrets(t *testing.T) {
	secretData := map[string][]byte{
		"che-admin-username": []byte("cheadmin"),
		"che-admin-password": []byte("password"),
	}
	secretData2 := map[string][]byte{
		"che-admin-username2": []byte("cheadmin2"),
		"che-admin-password2": []byte("password2"),
	}

	t.Run("one secret found", func(t *testing.T) {
		// given
		secret := test.CreateSecret("che-secret", test.MemberOperatorNs, secretData)
		cl := test.NewFakeClient(t, secret)

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		expected := map[string]string{
			"che-admin-username": "cheadmin",
			"che-admin-password": "password",
		}
		require.NoError(t, err)
		require.Equal(t, expected, secrets["che-secret"])
	})

	t.Run("two secrets found", func(t *testing.T) {
		// given
		secret := test.CreateSecret("che-secret", test.MemberOperatorNs, secretData)
		secret2 := test.CreateSecret("che-secret2", test.MemberOperatorNs, secretData2)
		cl := test.NewFakeClient(t, secret, secret2)

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		expected := map[string]string{
			"che-admin-username": "cheadmin",
			"che-admin-password": "password",
		}
		expected2 := map[string]string{
			"che-admin-username2": "cheadmin2",
			"che-admin-password2": "password2",
		}
		require.NoError(t, err)
		require.Equal(t, expected, secrets["che-secret"])
		require.Equal(t, expected2, secrets["che-secret2"])
	})

	t.Run("secrets from another namespace not listed", func(t *testing.T) {
		// given
		secret := test.CreateSecret("che-secret", test.MemberOperatorNs, secretData)
		secret.Namespace = "default"
		secret2 := test.CreateSecret("che-secret2", test.MemberOperatorNs, secretData2)
		secret2.Namespace = "default"
		cl := test.NewFakeClient(t, secret, secret2)

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		require.NoError(t, err)
		require.Empty(t, secrets)
	})

	t.Run("service account secrets are not listed", func(t *testing.T) {
		// given
		secret := test.CreateSecret("che-secret", test.MemberOperatorNs, secretData)
		secret.Annotations = map[string]string{
			"kubernetes.io/service-account.name": "default-something",
		}
		secret2 := test.CreateSecret("che-secret2", test.MemberOperatorNs, secretData2)
		secret2.Annotations = map[string]string{
			"kubernetes.io/service-account.name": "builder-something",
		}
		secret3 := test.CreateSecret("che-secret3", test.MemberOperatorNs, secretData2)
		secret3.Annotations = map[string]string{
			"kubernetes.io/service-account.name": "deployer-something",
		}
		cl := test.NewFakeClient(t, secret, secret2, secret3)

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		require.NoError(t, err)
		require.Empty(t, secrets)
	})

	t.Run("no secrets found", func(t *testing.T) {
		// given
		cl := test.NewFakeClient(t)

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		require.NoError(t, err)
		require.Empty(t, secrets)
	})

	t.Run("list secrets error", func(t *testing.T) {
		// given
		cl := test.NewFakeClient(t)
		cl.MockList = func(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
			return fmt.Errorf("list error")
		}

		// when
		secrets, err := LoadSecrets(cl, test.MemberOperatorNs)

		// then
		require.EqualError(t, err, "list error")
		require.Empty(t, secrets)
	})
}

func createConfigMap(name, namespace string, data map[string]string) *v1.ConfigMap { //nolint: unparam
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
	}
}
