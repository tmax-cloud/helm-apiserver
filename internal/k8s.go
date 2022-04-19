package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Namespace can retrieve current namespace
func Namespace() (string, error) {
	nsPath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if FileExists(nsPath) {
		// Running in k8s cluster
		nsBytes, err := ioutil.ReadFile(nsPath)
		if err != nil {
			return "", fmt.Errorf("could not read file %s", nsPath)
		}
		return string(nsBytes), nil
	}
	// Not running in k8s cluster (may be running locally)
	ns := os.Getenv("NAMESPACE")
	if ns == "" {
		ns = "template-system"
	}
	return ns, nil
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
