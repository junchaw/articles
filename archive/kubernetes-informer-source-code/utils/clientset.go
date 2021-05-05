package utils

import (
	"context"
	"fmt"
	"github.com/spongeprojects/magicconch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func MustClientset() kubernetes.Interface {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	magicconch.Must(err)

	clientset, err := kubernetes.NewForConfig(config)
	magicconch.Must(err)

	return clientset
}

func main() {
	clientset := MustClientset()

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	magicconch.Must(err)

	for _, namespace := range namespaces.Items {
		fmt.Println(namespace.Name)
	}
}
