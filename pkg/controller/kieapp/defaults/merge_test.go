package defaults

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	appsv1 "github.com/openshift/api/apps/v1"
	"github.com/sirupsen/logrus"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeServices(t *testing.T) {
	baseline, err := getConsole("trial", "test")
	assert.Nil(t, err)
	overwrite := baseline.DeepCopy()

	service1 := baseline.Services[0]
	service1.Labels["source"] = "baseline"
	service1.Labels["baseline"] = "true"
	service2 := service1.DeepCopy()
	service2.Name = service1.Name + "-2"
	service4 := service1.DeepCopy()
	service4.Name = service1.Name + "-4"
	baseline.Services = append(baseline.Services, *service2)
	baseline.Services = append(baseline.Services, *service4)

	service1b := overwrite.Services[0]
	service1b.Labels["source"] = "overwrite"
	service1b.Labels["overwrite"] = "true"
	service3 := service1b.DeepCopy()
	service3.Name = service1b.Name + "-3"
	service5 := service1b.DeepCopy()
	service5.Name = service1b.Name + "-4"
	annotations := service5.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
		service5.Annotations = annotations
	}
	service5.Annotations["delete"] = "true"
	overwrite.Services = append(overwrite.Services, *service3)
	overwrite.Services = append(overwrite.Services, *service5)

	merge(&baseline, overwrite)
	assert.Equal(t, 3, len(baseline.Services), "Expected 3 services")
	finalService1 := baseline.Services[0]
	finalService2 := baseline.Services[1]
	finalService3 := baseline.Services[2]
	assert.Equal(t, "true", finalService1.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalService1.Labels["overwrite"], "Expected the overwrite label to also be set as part of the merge")
	assert.Equal(t, "overwrite", finalService1.Labels["source"], "Expected the source label to have been overwritten by merge")
	assert.Equal(t, "true", finalService2.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "baseline", finalService2.Labels["source"], "Expected the source label to be baseline")
	assert.Equal(t, "true", finalService3.Labels["overwrite"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalService3.Labels["overwrite"], "Expected the overwrite label to be set")
	assert.Equal(t, "overwrite", finalService3.Labels["source"], "Expected the source label to be overwrite")
	assert.Equal(t, "test-rhpamcentr-2", finalService2.Name, "Second service name should end with -2")
	assert.Equal(t, "test-rhpamcentr-2", finalService2.Name, "Second service name should end with -3")
}

func TestMergeRoutes(t *testing.T) {
	baseline, err := getConsole("trial", "test")
	assert.Nil(t, err)
	overwrite := baseline.DeepCopy()

	route1 := baseline.Routes[0]
	route1.Labels["source"] = "baseline"
	route1.Labels["baseline"] = "true"
	route2 := route1.DeepCopy()
	route2.Name = route1.Name + "-2"
	route4 := route1.DeepCopy()
	route4.Name = route1.Name + "-4"
	baseline.Routes = append(baseline.Routes, *route2)
	baseline.Routes = append(baseline.Routes, *route4)

	route1b := overwrite.Routes[0]
	route1b.Labels["source"] = "overwrite"
	route1b.Labels["overwrite"] = "true"
	route3 := route1b.DeepCopy()
	route3.Name = route1b.Name + "-3"
	route5 := route1b.DeepCopy()
	route5.Name = route1b.Name + "-4"
	annotations := route5.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
		route5.Annotations = annotations
	}
	route5.Annotations["delete"] = "true"
	overwrite.Routes = append(overwrite.Routes, *route3)
	overwrite.Routes = append(overwrite.Routes, *route5)

	merge(&baseline, overwrite)
	assert.Equal(t, 3, len(baseline.Routes), "Expected 3 routes")
	finalRoute1 := baseline.Routes[0]
	finalRoute2 := baseline.Routes[1]
	finalRoute3 := baseline.Routes[2]
	assert.Equal(t, "true", finalRoute1.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalRoute1.Labels["overwrite"], "Expected the overwrite label to also be set as part of the merge")
	assert.Equal(t, "overwrite", finalRoute1.Labels["source"], "Expected the source label to have been overwritten by merge")
	assert.Equal(t, "true", finalRoute2.Labels["baseline"], "Expected the baseline label to be set")
	assert.Equal(t, "baseline", finalRoute2.Labels["source"], "Expected the source label to be baseline")
	assert.Equal(t, "true", finalRoute3.Labels["overwrite"], "Expected the baseline label to be set")
	assert.Equal(t, "true", finalRoute3.Labels["overwrite"], "Expected the overwrite label to be set")
	assert.Equal(t, "overwrite", finalRoute3.Labels["source"], "Expected the source label to be overwrite")
	assert.Equal(t, "test-rhpamcentr-2", finalRoute2.Name, "Second route name should end with -2")
	assert.Equal(t, "test-rhpamcentr-2", finalRoute2.Name, "Second route name should end with -3")
}

func getConsole(environment string, name string) (v1.CustomObject, error) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: environment,
		},
	}

	env, _, err := GetEnvironment(cr)
	if err != nil {
		return v1.CustomObject{}, err
	}
	return env.Console, nil
}

func TestMergeServerDeploymentConfigs(t *testing.T) {
	var prodEnv v1.Environment
	err := getParsedTemplate("envs/production-lite.yaml", "prod", &prodEnv)
	assert.Nil(t, err, "Error: %v", err)

	var servers v1.CustomObject
	err = getParsedTemplate("common/server.yaml", "prod", &servers)
	assert.Nil(t, err, "Error: %v", err)

	baseEnvCount := len(servers.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)
	prodEnvCount := len(prodEnv.Servers[0].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env)

	mergedDCs := mergeDeploymentConfigs(servers.DeploymentConfigs, prodEnv.Servers[0].DeploymentConfigs)

	assert.NotNil(t, mergedDCs, "Must have encountered an error, merged DCs should not be null")
	assert.Len(t, mergedDCs, 2, "Expect 2 deployment descriptors but got %v", len(mergedDCs))

	mergedEnvCount := len(mergedDCs[0].Spec.Template.Spec.Containers[0].Env)
	assert.True(t, mergedEnvCount > baseEnvCount, "Merged DC should have a higher number of environment variables than the base server")
	assert.True(t, mergedEnvCount > prodEnvCount, "Merged DC should have a higher number of environment variables than the server")

	assert.Len(t, mergedDCs[0].Spec.Template.Spec.Containers[0].Ports, 4, "Expecting 4 ports")
}

func TestMergeAuthoringServer(t *testing.T) {
	var prodEnv v1.Environment
	err := getParsedTemplate("testdata/envs/authoring-lite.yaml", "prod", &prodEnv)
	assert.Nil(t, err, "Error: %v", err)
	var servers, expected v1.CustomObject
	err = getParsedTemplate("common/server.yaml", "prod", &servers)
	assert.Nil(t, err, "Error: %v", err)

	merge(&servers, &prodEnv.Servers[0])

	err = getParsedTemplate("testdata/expected/authoring.yaml", "fake", &expected)
	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, &expected, &servers)
}

func TestMergeAuthoringPostgresServer(t *testing.T) {
	var prodEnv v1.Environment
	err := getParsedTemplate("testdata/envs/authoring-postgres-lite.yaml", "prod", &prodEnv)
	assert.Nil(t, err, "Error: %v", err)
	var servers, expected v1.CustomObject
	err = getParsedTemplate("common/server.yaml", "prod", &servers)
	assert.Nil(t, err, "Error: %v", err)

	merge(&servers, &prodEnv.Servers[0])

	err = getParsedTemplate("testdata/expected/authoring-postgres.yaml", "fake", &expected)
	var d, _ = yaml.Marshal(&servers)
	fmt.Printf("########MERGED\n%s", d)

	marshalledExpected, _ := yaml.Marshal(&expected)
	fmt.Printf("--------Expected: \n%s", marshalledExpected)
	fmt.Printf("are equal: %v\n", reflect.DeepEqual(&expected, &servers))
	assert.Nil(t, err, "Error: %v", err)
	assert.Equal(t, &expected, &servers)
}

func TestMergeDeploymentconfigs(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}

	overwrite := []appsv1.DeploymentConfig{
		*buildDC("overwrite-dc2"),
		*buildDC("dc1"),
	}
	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, overwrite[0], results[1])
	assert.Equal(t, overwrite[1], results[0])
}

func TestMergeDeploymentconfigs_Metadata(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		appsv1.DeploymentConfig{
			ObjectMeta: *buildObjectMeta("dc1-dc"),
		},
	}
	baseline[0].ObjectMeta.Labels["foo"] = "replace me"
	baseline[0].ObjectMeta.Labels["john"] = "doe"
	overwrite[0].ObjectMeta.Labels["foo"] = "replaced"
	overwrite[0].ObjectMeta.Labels["ping"] = "pong"

	baseline[0].ObjectMeta.Annotations["foo"] = "replace me"
	baseline[0].ObjectMeta.Annotations["john"] = "doe"
	overwrite[0].ObjectMeta.Annotations["foo"] = "replaced"
	overwrite[0].ObjectMeta.Annotations["ping"] = "pong"

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, "replaced", results[0].ObjectMeta.Labels["foo"])
	assert.Equal(t, "doe", results[0].ObjectMeta.Labels["john"])
	assert.Equal(t, "pong", results[0].ObjectMeta.Labels["ping"])

	assert.Equal(t, "replaced", results[0].ObjectMeta.Annotations["foo"])
	assert.Equal(t, "doe", results[0].ObjectMeta.Annotations["john"])
	assert.Equal(t, "pong", results[0].ObjectMeta.Annotations["ping"])
}

func TestMergeDeploymentconfigs_TemplateMetadata(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	baseline[0].Spec.Template.ObjectMeta.Labels["foo"] = "replace me"
	baseline[0].Spec.Template.ObjectMeta.Labels["john"] = "doe"
	overwrite[0].Spec.Template.ObjectMeta.Labels["foo"] = "replaced"
	overwrite[0].Spec.Template.ObjectMeta.Labels["ping"] = "pong"

	baseline[0].Spec.Template.ObjectMeta.Annotations["foo"] = "replace me"
	baseline[0].Spec.Template.ObjectMeta.Annotations["john"] = "doe"
	overwrite[0].Spec.Template.ObjectMeta.Annotations["foo"] = "replaced"
	overwrite[0].Spec.Template.ObjectMeta.Annotations["ping"] = "pong"

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, "replaced", results[0].Spec.Template.ObjectMeta.Labels["foo"])
	assert.Equal(t, "doe", results[0].Spec.Template.ObjectMeta.Labels["john"])
	assert.Equal(t, "pong", results[0].Spec.Template.ObjectMeta.Labels["ping"])

	assert.Equal(t, "replaced", results[0].Spec.Template.ObjectMeta.Annotations["foo"])
	assert.Equal(t, "doe", results[0].Spec.Template.ObjectMeta.Annotations["john"])
	assert.Equal(t, "pong", results[0].Spec.Template.ObjectMeta.Annotations["ping"])
}

func TestMergeDeploymentconfigs_Spec(t *testing.T) {
	baseline := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite := []appsv1.DeploymentConfig{
		*buildDC("dc1"),
	}
	overwrite[0].Spec.Strategy.Type = "Other Strategy"
	overwrite[0].Spec.Triggers[0] = appsv1.DeploymentTriggerPolicy{
		Type: "ImageChange",
		ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
			From: corev1.ObjectReference{
				Name: "other-image:future",
			},
		},
	}
	overwrite[0].Spec.Triggers = append(overwrite[0].Spec.Triggers, appsv1.DeploymentTriggerPolicy{
		Type: "ConfigChange",
	})

	results := mergeDeploymentConfigs(baseline, overwrite)

	assert.Equal(t, appsv1.DeploymentStrategyType("Other Strategy"), results[0].Spec.Strategy.Type)
	assert.Equal(t, 2, len(results[0].Spec.Triggers))
	assert.Equal(t, "openshift", results[0].Spec.Triggers[0].ImageChangeParams.From.Namespace)
	assert.Equal(t, "other-image:future", results[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Equal(t, appsv1.DeploymentTriggerType("ConfigChange"), results[0].Spec.Triggers[1].Type)
}

func getParsedTemplate(filename string, name string, object interface{}) error {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
	}
	envTemplate := getEnvTemplate(cr)

	yamlBytes, err := loadYaml(filename, envTemplate)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlBytes, object)
	if err != nil {
		logrus.Error(err)
	}
	return nil
}

func buildDC(name string) *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: *buildObjectMeta(name + "-dc"),
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				{
					Type: "ImageChange",
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							name + "-container-1",
						},
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Namespace: "openshift",
							Name:      "rhpam70-kieserver:latest",
						},
					},
				},
			},
			Replicas: 3,
			Selector: map[string]string{
				"deploymentConfig": name,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: *buildObjectMeta(name + "-tplt"),
				Spec: corev1.PodSpec{
					ServiceAccountName: name + "test-sa",
					Containers: []corev1.Container{
						{
							Name:  name + "container",
							Image: "image-" + name,
							Env: []corev1.EnvVar{
								{
									Name:  name + "-env-1",
									Value: name + "-val-1",
								},
								{
									Name:  name + "-env-2",
									Value: name + "-val-2",
								},
								{
									Name: name + "-env-3",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: name + "-configmap",
											},
											Key: name + "-configmap-key",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"memory": *resource.NewQuantity(1, "Mi"),
									"cpu":    *resource.NewQuantity(1, ""),
								},
								Requests: corev1.ResourceList{
									"memory": *resource.NewQuantity(2, "Mi"),
									"cpu":    *resource.NewQuantity(2, ""),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      name + "-volume-mount-1",
									MountPath: "/etc/kieserver/" + name + "/path1",
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          name + "-port1",
									Protocol:      "TCP",
									ContainerPort: 9090,
								},
								{
									Name:          name + "-port2",
									Protocol:      "TCP",
									ContainerPort: 8443,
								},
							},
							LivenessProbe:  buildProbe(name+"-liveness", 30, 2),
							ReadinessProbe: buildProbe(name+"-readiness", 60, 4),
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: name + "-emptydir-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: name + "-secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: name + "-secret",
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildObjectMeta(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: name + "-ns",
		Labels: map[string]string{
			name + ".label1": name + "-labelValue1",
			name + ".label2": name + "-labelValue2",
		},
		Annotations: map[string]string{
			name + ".annotation1": name + "-annValue1",
			name + ".annotation2": name + "-annValue2",
		},
	}
}

func buildProbe(name string, delay, timeout int32) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/" + name,
					"-c",
					name,
				},
			},
		},
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
	}
}