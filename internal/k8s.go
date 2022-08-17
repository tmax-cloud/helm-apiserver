package internal

import (
	"context"
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nsEnv = "NAMESPACE"
const nsFilePathDefault = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

var nsFilePath = nsFilePathDefault

// Namespace can retrieve current namespace
func Namespace() string {
	if nsBytes, err := ioutil.ReadFile(nsFilePath); err == nil {
		return string(nsBytes)
	}

	// Fallback to env, default values
	// Not running in k8s cluster (maybe running locally)
	ns := os.Getenv(nsEnv)
	if ns == "" {
		ns = "helm-ns"
	}
	return ns
}

func GetServiceAccount(c client.Client, name types.NamespacedName) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}

	if err := c.Get(context.TODO(), name, serviceAccount); err != nil {
		return nil, err
	}
	return serviceAccount, nil
}

func GetSecret(c client.Client, name types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	if err := c.Get(context.TODO(), name, secret); err != nil {
		return nil, err
	}
	return secret, nil
}
