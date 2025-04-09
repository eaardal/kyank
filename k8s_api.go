package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sApi struct {
	clientset *kubernetes.Clientset
	namespace string
}

func newK8sApi(namespace string) *k8sApi {
	return &k8sApi{
		namespace: namespace,
	}
}

func (k *k8sApi) init() error {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes clientset: %v", err)
	}

	k.clientset = clientset
	return nil
}

func (k *k8sApi) getPod(ctx context.Context, podId string) (*corev1.Pod, error) {
	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podId, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s in namespace %s: %v", podId, k.namespace, err)
	}
	return pod, nil
}

func (k *k8sApi) getSecretValue(ctx context.Context, secretName string, key string) (string, error) {
	secret, err := k.clientset.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch secret %s: %v", secretName, err)
	}

	decoded, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", key, secretName)
	}

	return string(decoded), nil
}
