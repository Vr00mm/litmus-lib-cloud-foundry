package k8s

import (
//	"context"
	"fmt"
	"os"

//	corev1r "k8s.io/api/core/v1"

	metav1r "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd"
)

func connectKubeApi() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("❌ Cannot create config from incluster: " + err.Error() + "\n")
		os.Exit(1)
	}
	// creates the clientset
	// kubeConfigPath := "/home/erse9980/.kube/configs/config_pov-sddc"
	// fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	// config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	// if err != nil {
	// 	fmt.Printf("error getting Kubernetes config: %v\n", err)
	// 	os.Exit(1)
	// }
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("❌ Cannot create clientset: " + err.Error() + "\n")
		os.Exit(1)
	}
	return clientset
}

func GetSecretAsMap(name string, namespace string) (map[string]string, error) {
        clientset := connectKubeApi()

	var getOption metav1r.GetOptions
	secret, err := clientset.CoreV1().Secrets(namespace).Get(name, getOption)
	if err != nil {
                fmt.Printf("error")
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		result[k] = string(v)
	}

	return result, nil
}

func LoadSecretAsEnv(name string, namespace string) error {
	secret, err := GetSecretAsMap(name, namespace)
        if err != nil {
                return err
        }

        for k, v := range secret {
                os.Setenv(k, v)
        }

	return nil

}
