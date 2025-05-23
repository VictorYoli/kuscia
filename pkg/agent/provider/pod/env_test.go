// Copyright 2023 Ant Group Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pod

import (
	"context"
	"sort"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgcontainer "github.com/secretflow/kuscia/pkg/agent/container"

	resourcetest "github.com/secretflow/kuscia/pkg/agent/resource/testing"
	testutil "github.com/secretflow/kuscia/tests/util"
)

const (
	// defaultEventRecorderBufferSize is the default buffer size to use when creating fake event recorders.
	defaultEventRecorderBufferSize = 5
	// envVarName1 is a string that can be used as the name of an environment value.
	envVarName1 = "FOO"
	// envVarValue1 is a string meant to be used as the value of the "envVarName1" environment value.
	envVarValue1 = "foo_value"
	// envVarName2 is a string that can be used as the name of an environment value.
	envVarName2 = "BAR"
	// envVarValue2 is a string meant to be used as the value of the "envVarName2" environment value.
	envVarValue2 = "bar_value"
	// envVarName12 is a key that can be used as the name of an environment variable.
	envVarName12 = "FOOBAR"
	// envVarName3 is a string that can be used as the name of an environment value.
	envVarName3 = "CHO"
	// envVarName4 is a string that can be used as the name of an environment value.
	envVarName4 = "CAR"
	// invalidKey1 is a key that cannot be used as the name of an environment variable (since it starts with a digit).
	invalidKey1 = "1INVALID"
	// invalidKey2 is a key that cannot be used as the name of an environment variable (since it starts with a digit).
	invalidKey2 = "2INVALID"
	// invalidKey3 is a key that cannot be used as the name of an environment variable (since it starts with a digit).
	invalidKey3 = "3INVALID"
	// keyFoo is a key that can be used as the name of an environment variable.
	keyFoo = "FOO"
	// keyBar is a key that can be used as the name of an environment variable.
	keyBar = "BAR"
	// keyBaz is a key that can be used as the name of an environment variable.
	keyBaz = "BAZ"
	// namespace is the namespace to which mock resources used in the tests belong.
	namespace = "foo"
	// prefixConfigMap1 is the prefix used in ".envFrom" fields that reference "config-map-1".
	prefixConfigMap1 = "FROM_CONFIGMAP_1_"
	// prefixConfigMap2 is the prefix used in ".envFrom" fields that reference "config-map-2".
	prefixConfigMap2 = "FROM_CONFIGMAP_2_"
	// prefixSecret1 is the prefix used in ".envFrom" fields that reference "secret-1".
	prefixSecret1 = "FROM_SECRET_1_"
	// prefixSecret1 is the prefix used in ".envFrom" fields that reference "secret-1".
	prefixSecret2 = "FROM_SECRET_2_"
)

var (
	// bFalse represents the "false" value.
	// Used so we can take its address when a pointer to a bool is required.
	bFalse = false
	// bFalse represents the "true" value.
	// Used so we can take its address when a pointer to a bool is required.
	bTrue = true
	// configMap1 is a configmap containing a single key, valid as the name of an environment variable.
	configMap1 = testutil.FakeConfigMap(namespace, "configmap-1", map[string]string{
		keyFoo: "__foo__",
	})
	// configMap2 is a configmap containing a single key, valid as the name of an environment variable.
	configMap2 = testutil.FakeConfigMap(namespace, "configmap-2", map[string]string{
		keyBar: "__bar__",
	})
	// configMap3 is a configmap containing a single key, valid as the name of an environment variable.
	configMap3 = testutil.FakeConfigMap(namespace, "configmap-2", map[string]string{
		keyFoo: "__foo__",
		keyBar: "__bar__",
	})
	// invalidConfigMap1 is a configmap containing two keys, one of which is invalid as the name of an environment variable.
	invalidConfigMap1 = testutil.FakeConfigMap(namespace, "invalid-configmap-1", map[string]string{
		keyFoo:      "__foo__",
		invalidKey1: "will-be-skipped",
		invalidKey2: "will-be-skipped",
	})
	// secret1 is a secret containing a single key, valid as the name of an environment variable.
	secret1 = testutil.FakeSecret(namespace, "secret-1", map[string]string{
		keyBaz: "__baz__",
	})
	// secret2 is a secret containing a single key, valid as the name of an environment variable.
	secret2 = testutil.FakeSecret(namespace, "secret-2", map[string]string{
		keyFoo: "__foo__",
	})
	// invalidSecret1 is a secret containing two keys, one of which is invalid as the name for an environment variable.
	invalidSecret1 = testutil.FakeSecret(namespace, "invalid-secret-1", map[string]string{
		invalidKey3: "will-be-skipped",
		keyBaz:      "__baz__",
	})
)

// TestPopulatePodWithInitContainersUsingEnv populates the environment of a pod with four containers (two init containers, two containers) using ".env".
// Then, it checks that the resulting environment for each container contains the expected environment variables.
func TestPopulatePodWithInitContainersUsingEnv(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, configMap1, configMap2, secret1, secret2)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// Create a pod object having two init containers and two containers.
	// The containers' environment is to be populated both from directly-provided values and from keys in two configmaps and two secrets.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name:  envVarName1,
							Value: envVarValue1,
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap1.Name,
									},
									Key:      keyFoo,
									Optional: nil,
								},
							},
						},
					},
				},
				{
					Env: []corev1.EnvVar{
						{
							Name:  envVarName1,
							Value: envVarValue1,
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secret1.Name,
									},
									Key:      keyBaz,
									Optional: &bFalse,
								},
							},
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name:  envVarName1,
							Value: envVarValue1,
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap1.Name,
									},
									Key:      keyFoo,
									Optional: nil,
								},
							},
						},
					},
				},
				{
					Env: []corev1.EnvVar{
						{
							Name:  envVarName1,
							Value: envVarValue1,
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secret1.Name,
									},
									Key:      keyBaz,
									Optional: &bFalse,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	tests := []struct {
		container *corev1.Container
		envVar    []pkgcontainer.EnvVar
	}{
		{
			container: &pod.Spec.InitContainers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: envVarValue1,
				},
				{
					Name:  envVarName2,
					Value: configMap1.Data[keyFoo],
				},
			},
		},
		{
			container: &pod.Spec.InitContainers[1],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: envVarValue1,
				},
				{
					Name:  envVarName2,
					Value: string(secret1.Data[keyBaz]),
				},
			},
		},
		{
			container: &pod.Spec.Containers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: envVarValue1,
				},
				{
					Name:  envVarName2,
					Value: configMap1.Data[keyFoo],
				},
			},
		},
		{
			container: &pod.Spec.Containers[1],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: envVarValue1,
				},
				{
					Name:  envVarName2,
					Value: string(secret1.Data[keyBaz]),
				},
			},
		},
	}

	for i, tt := range tests {
		res, err := populateContainerEnvironment(context.Background(), pod, tt.container, rm, er)
		assert.NoError(t, err, "Test %d", i)
		assert.ElementsMatch(t, tt.envVar, res, "Test %d", i)
	}
}

// TestPopulatePodWithInitContainersUsingEnv populates the environment of a pod with four containers (two init containers, two containers) using ".env".
// Then, it checks that the resulting environment for each container contains the expected environment variables.
func TestPopulatePodWithInitContainersUsingEnvWithFieldRef(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, configMap1, configMap2, secret1, secret2)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// Create a pod object having two init containers and two containers.
	// The containers' environment is to be populated both from directly-provided values and from keys in two configmaps and two secrets.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
			Labels: map[string]string{
				"zone":    "us-est-coast",
				"cluster": "test-cluster1",
				"rack":    "rack-22",
			},
			Annotations: map[string]string{
				"build":   "two",
				"builder": "john-doe",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:           "namenode",
			ServiceAccountName: "serviceaccount",
			InitContainers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: envVarName1,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "spec.nodeName",
								},
							},
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.labels",
								},
							},
						},
						{
							Name: envVarName3,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.annotations",
								},
							},
						},
						{
							Name: envVarName4,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "spec.serviceAccountName",
								},
							},
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: envVarName1,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "spec.nodeName",
								},
							},
						},
						{
							Name: envVarName2,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.labels",
								},
							},
						},
						{
							Name: envVarName3,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "metadata.annotations",
								},
							},
						},
						{
							Name: envVarName4,
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath:  "spec.serviceAccountName",
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	tests := []struct {
		container *corev1.Container
		envVar    []pkgcontainer.EnvVar
	}{
		{
			container: &pod.Spec.InitContainers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: "namenode",
				},
				{
					Name:  envVarName2,
					Value: "cluster=\"test-cluster1\"\nrack=\"rack-22\"\nzone=\"us-est-coast\"",
				},
				{
					Name:  envVarName3,
					Value: "build=\"two\"\nbuilder=\"john-doe\"",
				},
				{
					Name:  envVarName4,
					Value: "serviceaccount",
				},
			},
		},
		{
			container: &pod.Spec.Containers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  envVarName1,
					Value: "namenode",
				},
				{
					Name:  envVarName2,
					Value: "cluster=\"test-cluster1\"\nrack=\"rack-22\"\nzone=\"us-est-coast\"",
				},
				{
					Name:  envVarName3,
					Value: "build=\"two\"\nbuilder=\"john-doe\"",
				},
				{
					Name:  envVarName4,
					Value: "serviceaccount",
				},
			},
		},
	}

	for i, tt := range tests {
		res, err := populateContainerEnvironment(context.Background(), pod, tt.container, rm, er)
		assert.NoError(t, err, "Test %d", i)
		assert.ElementsMatch(t, tt.envVar, res, "Test %d", i)
	}
}

// TestPopulatePodWithInitContainersUsingEnvFrom populates the environment of a pod with four containers (two init containers, two containers) using ".envFrom".
// Then, it checks that the resulting environment for each container contains the expected environment variables.
func TestPopulatePodWithInitContainersUsingEnvFrom(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, configMap1, configMap2, secret1, secret2)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// Create a pod object having two init containers and two containers.
	// The containers' environment is to be populated from two configmaps and two secrets.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: prefixConfigMap1,
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMap1.Name,
								},
							},
						},
					},
				},
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: prefixConfigMap2,
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMap2.Name,
								},
							},
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: prefixSecret1,
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secret1.Name,
								},
							},
						},
					},
				},
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: prefixSecret2,
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secret2.Name,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	tests := []struct {
		container *corev1.Container
		envVar    []pkgcontainer.EnvVar
	}{
		{
			container: &pod.Spec.InitContainers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  prefixConfigMap1 + keyFoo,
					Value: configMap1.Data[keyFoo],
				},
			},
		},
		{
			container: &pod.Spec.InitContainers[1],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  prefixConfigMap2 + keyBar,
					Value: configMap2.Data[keyBar],
				},
			},
		},
		{
			container: &pod.Spec.Containers[0],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  prefixSecret1 + keyBaz,
					Value: string(secret1.Data[keyBaz]),
				},
			},
		},
		{
			container: &pod.Spec.Containers[1],
			envVar: []pkgcontainer.EnvVar{
				{
					Name:  prefixSecret2 + keyFoo,
					Value: string(secret2.Data[keyFoo]),
				},
			},
		},
	}

	for i, tt := range tests {
		res, err := populateContainerEnvironment(context.Background(), pod, tt.container, rm, er)
		assert.NoError(t, err, "Test %d", i)
		assert.ElementsMatch(t, tt.envVar, res, "Test %d", i)
	}
}

// TestEnvFromTwoConfigMapsAndOneSecret populates the environment of a container from two configmaps and one secret.
// Then, it checks that the resulting environment contains all the expected environment variables and values.
func TestEnvFromTwoConfigMapsAndOneSecret(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, configMap1, configMap2, secret1)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// Create a pod object having a single container.
	// The container's environment is to be populated from two configmaps and one secret.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: prefixConfigMap1,
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMap1.Name,
								},
							},
						},
						{
							Prefix: prefixConfigMap2,
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMap2.Name,
								},
							},
						},
						{
							Prefix: prefixSecret1,
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secret1.Name,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the container's environment.
	res, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.NoError(t, err)

	assert.ElementsMatch(t, []pkgcontainer.EnvVar{
		{
			Name:  prefixConfigMap1 + keyFoo,
			Value: configMap1.Data[keyFoo],
		},
		{
			Name:  prefixConfigMap2 + keyBar,
			Value: configMap2.Data[keyBar],
		},
		{
			Name:  prefixSecret1 + keyBaz,
			Value: string(secret1.Data[keyBaz]),
		},
	}, res)

	// Make sure that no events have been recorded, as the configmaps and secrets are valid.
	assert.Len(t, er.Events, 0)
}

// TestEnvFromConfigMapAndSecretWithInvalidKeys populates the environment of a container from a configmap and a secret containing invalid keys.
// Then, it checks that the resulting environment contains all the expected environment variables and values, and that the invalid keys have been skipped.
func TestEnvFromConfigMapAndSecretWithInvalidKeys(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, invalidConfigMap1, invalidSecret1)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// Create a pod object having a single container.
	// The container's environment is to be populated from a configmap and a secret, both of which have some invalid keys.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: invalidConfigMap1.Name,
								},
							},
						},
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: invalidSecret1.Name,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	res, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.NoError(t, err)

	// Make sure that the container's environment has two variables (corresponding to the single valid key in both the configmap and the secret).
	assert.ElementsMatch(t, []pkgcontainer.EnvVar{
		{
			Name:  keyFoo,
			Value: invalidConfigMap1.Data[keyFoo],
		},
		{
			Name:  keyBaz,
			Value: string(invalidSecret1.Data[keyBaz]),
		},
	}, res)

	// Make sure that two events have been received (one for the configmap and one for the secret).
	assert.Len(t, er.Events, 2)

	// Grab the first event (which should correspond to the configmap) and make sure it has the correct reason and message.
	event1 := <-er.Events
	assert.Contains(t, event1, ReasonInvalidEnvironmentVariableNames)
	assert.Contains(t, event1, invalidKey1)
	assert.Contains(t, event1, invalidKey2)

	// Grab the second event (which should correspond to the secret) and make sure it has the correct reason and message.
	event2 := <-er.Events
	assert.Contains(t, event2, ReasonInvalidEnvironmentVariableNames)
	assert.Contains(t, event2, invalidKey3)
}

// TestEnvOverridesEnvFrom populates the environment of a container from a configmap, and from another configmap's key with a "conflicting" key.
// Then, it checks that the value of the "conflicting" key has been correctly overridden.
func TestEnvOverridesEnvFrom(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace, configMap3)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	// override will override the value of "keyFoo" from "configMap3".
	override := "__override__"

	// Create a pod object having a single container.
	// The container's environment is to be populated from a configmap, and later overridden with a value provided directly.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMap3.Name,
								},
							},
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  keyFoo, // One of the keys in configMap3.
							Value: override,
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	res, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.NoError(t, err)

	// Make sure that the container's environment has two variables (corresponding to the single valid key in both the configmap and the secret).
	assert.ElementsMatch(t, []pkgcontainer.EnvVar{
		{
			Name:  keyFoo,
			Value: override,
		},
		{
			Name:  keyBar,
			Value: configMap3.Data[keyBar],
		},
	}, res)

	// Make sure that two events have been received (one for the configmap and one for the secret).
	assert.Len(t, er.Events, 0)
}

var sortOpt = gocmp.Transformer("Sort", sortEnv)

func sortEnv(in []corev1.EnvVar) []corev1.EnvVar {
	out := append([]corev1.EnvVar(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// TestEnvFromInexistentConfigMaps populates the environment of a container from two configmaps (one of them optional) that do not exist.
// Then, it checks that the expected events have been recorded.
func TestEnvFromInexistentConfigMaps(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	missingConfigMap1Name := "missing-config-map-1"
	missingConfigMap2Name := "missing-config-map-2"

	// Create a pod object having a single container.
	// The container's environment is to be populated from two configmaps that do not exist.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: missingConfigMap1Name,
								},
								// The configmap reference is optional.
								Optional: &bTrue,
							},
						},
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: missingConfigMap2Name,
								},
								// The configmap reference is mandatory.
								Optional: &bFalse,
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	_, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.ErrorContains(t, err, "not found")

	assert.Len(t, er.Events, 2)

	// Grab the first event (which should correspond to the configmap) and make sure it has the correct reason and message.
	event1 := <-er.Events
	assert.Contains(t, event1, ReasonOptionalConfigMapNotFound)
	assert.Contains(t, event1, missingConfigMap1Name)
	event2 := <-er.Events
	assert.Contains(t, event2, ReasonMandatoryConfigMapNotFound)
	assert.Contains(t, event2, missingConfigMap2Name)
}

func TestEnvFromInexistentSecrets(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	missingSecret1Name := "missing-secret-1"
	missingSecret2Name := "missing-secret-2"

	// Create a pod object having a single container.
	// The container's environment is to be populated from two secrets that do not exist.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: missingSecret1Name,
								},
								// The secret reference is optional.
								Optional: &bTrue,
							},
						},
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: missingSecret2Name,
								},
								// The secret reference is mandatory.
								Optional: &bFalse,
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	_, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.ErrorContains(t, err, "not found")

	assert.Len(t, er.Events, 2)

	// Grab the first event (which should correspond to the configmap) and make sure it has the correct reason and message.
	event1 := <-er.Events
	assert.Contains(t, event1, ReasonOptionalSecretNotFound)
	assert.Contains(t, event1, missingSecret1Name)
	event2 := <-er.Events
	assert.Contains(t, event2, ReasonMandatorySecretNotFound)
	assert.Contains(t, event2, missingSecret2Name)
}

// TestEnvReferencingInexistentConfigMapKey tries populates the environment of a container using a keys from a configmaps that does not exist.
// Then, it checks that the expected event has been recorded.
func TestEnvReferencingInexistentConfigMapKey(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	missingConfigMapName := "missing-config-map"

	// Create a pod object having a single container and referencing a key from a configmap that does not exist.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: "envvar",
							ValueFrom: &corev1.EnvVarSource{
								ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: missingConfigMapName,
									},
									Key: "key",
									// A nil value of optional means "mandatory", hence we should expect "populateEnvironmentVariables" to return an error.
									Optional: nil,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	_, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.ErrorContains(t, err, "not found")

	assert.Len(t, er.Events, 1)

	// Grab the first event (which should correspond to the configmap) and make sure it has the correct reason and message.
	event1 := <-er.Events
	assert.Contains(t, event1, ReasonMandatoryConfigMapNotFound)
	assert.Contains(t, event1, missingConfigMapName)
}

// TestEnvReferencingInexistentSecretKey tries populates the environment of a container using a keys from a secret that does not exist.
// Then, it checks that the expected event has been recorded.
func TestEnvReferencingInexistentSecretKey(t *testing.T) {
	rm := resourcetest.FakeResourceManager(namespace)
	er := testutil.FakeEventRecorder(defaultEventRecorderBufferSize)

	missingSecretName := "missing-secret"

	// Create a pod object having a single container and referencing a key from a secret that does not exist.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "pod-0",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name: "envvar",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: missingSecretName,
									},
									Key: "key",
									// A nil value of optional means "mandatory", hence we should expect "populateEnvironmentVariables" to return an error.
									Optional: nil,
								},
							},
						},
					},
				},
			},
			EnableServiceLinks: &bFalse,
		},
	}

	// Populate the pods's environment.
	_, err := populateContainerEnvironment(context.Background(), pod, &pod.Spec.Containers[0], rm, er)
	assert.ErrorContains(t, err, "not found")

	assert.Len(t, er.Events, 1)

	// Grab the first event (which should correspond to the configmap) and make sure it has the correct reason and message.
	event1 := <-er.Events
	assert.Contains(t, event1, ReasonMandatorySecretNotFound)
	assert.Contains(t, event1, missingSecretName)
}
