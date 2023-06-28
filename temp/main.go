// package main

// import (
// 	"context"
// 	"flag"
// 	"fmt"

// 	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/tools/clientcmd"
// )

// func main() {
// 	kubeconfig := flag.String("kubeconfig", "", "path to the kubeconfig file")
// 	flag.Parse()

// 	// Load the kubeconfig file
// 	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Create the Kubernetes clientset
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// List all the ValidatingWebhookConfiguration objects
// 	validatingWebhookConfigs, err := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Iterate over the ValidatingWebhookConfiguration objects
// 	for _, validatingWebhookConfig := range validatingWebhookConfigs.Items {
// 		// Get the admission policies
// 		admissionPolicies := getAdmissionPolicies(validatingWebhookConfig.Webhooks)
// 		fmt.Printf("ValidatingWebhookConfiguration: %s\n", validatingWebhookConfig.Name)
// 		fmt.Printf("Admission Policies: %v\n", admissionPolicies)
// 		fmt.Println()
// 	}
// }

// // getAdmissionPolicies extracts the admission policies from the given list of webhooks
// func getAdmissionPolicies(webhooks []admissionregistrationv1.ValidatingWebhook) []string {
// 	admissionPolicies := make([]string, 0)
// 	for _, webhook := range webhooks {
// 		admissionPolicies = append(admissionPolicies, webhook.AdmissionReviewVersions...)
// 	}
// 	return admissionPolicies
// }
