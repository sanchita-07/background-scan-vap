package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	// "github.com/kyverno/kyverno/pkg/clients/dclient"
	kyverno "github.com/kyverno/kyverno/pkg/clients/dclient"
	// config "github.com/kyverno/kyverno/pkg/config"
	"k8s.io/api/admissionregistration/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

type Interface interface {
	// GetKubeClient provides typed kube client
	GetKubeClient() kubernetes.Interface
	// GetDynamicInterface fetches underlying dynamic interface
	GetDynamicInterface() dynamic.Interface
	// Discovery return the discovery client implementation
	Discovery() IDiscovery
}

// Client enables interaction with k8 resource
type client struct {
	dyn   dynamic.Interface
	disco IDiscovery
	rest  rest.Interface
	kube  kubernetes.Interface
}

type ApplyPolicyConfig struct {
	Policy                    kyvernov1.PolicyInterface
	ValidatingAdmissionPolicy v1alpha1.ValidatingAdmissionPolicy
	Resource                  *unstructured.Unstructured
	// Variables                 map[string]interface{}
	PolicyReport              bool
	// NamespaceSelectorMap      map[string]map[string]string
	// Stdin                     bool
	// PrintPatchResource        bool
	// RuleToCloneSourceResource map[string]string
	Client kyverno.Interface
}

// NewClient creates new instance of client
func NewClient(
	ctx context.Context,
	dyn dynamic.Interface,
	kube kubernetes.Interface,
	resync time.Duration,
) (Interface, error) {
	disco := kube.Discovery()
	client := client{
		dyn:  dyn,
		kube: kube,
		rest: disco.RESTClient(),
	}
	// Set discovery client
	discoveryClient := &serverResources{
		cachedClient: memory.NewMemCacheClient(disco),
	}
	// client will invalidate registered resources cache every x seconds,
	// As there is no way to identify if the registered resource is available or not
	// we will be invalidating the local cache, so the next request get a fresh cache
	// If a resource is removed then and cache is not invalidate yet, we will not detect the removal
	// but the re-sync shall re-evaluate
	go discoveryClient.Poll(ctx, resync)
	client.SetDiscovery(discoveryClient)
	return &client, nil
}

// BackgroundScanner is responsible for performing background scanning of ValidatingAdmissionPolicy resources in Kyverno
type BackgroundScanner struct {
	kyvernoClient *kyverno.Client
	eventRecorder record.EventRecorder
}

// NewBackgroundScanner creates a new instance of BackgroundScanner
func NewBackgroundScanner(kyvernoClient *kyverno.Client, eventRecorder record.EventRecorder) *BackgroundScanner {
	return &BackgroundScanner{
		kyvernoClient: kyvernoClient,
		eventRecorder: eventRecorder,
	}
}

func main() {
	// Parse command line flags
	kubeconfig := flag.String("kubeconfig", "", "Path to the kubeconfig file")
	flag.Parse()

	// Build Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	// Create event recorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	eventRecorder := eventBroadcaster.NewRecorder(runtime.NewScheme(), v1.EventSource{Component: "kyverno-background-scan"})

	// Create Kyverno client
	kyvernoClient, err := kyverno.NewClient(config)
	if err != nil {
		klog.Fatalf("Failed to create Kyverno client: %v", err)
	}

	// Create Kyverno background scanner
	backgroundScanner := kyverno.NewBackgroundScanner(kyvernoClient, eventRecorder)

	// Start background scanning
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := backgroundScanner.Start(ctx)
		if err != nil {
			klog.Errorf("Background scanning error: %v", err)
		}
	}()

	// Handle termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalCh:
		klog.Info("Termination signal received, stopping background scanning...")
		cancel()
	}

	// Wait for background scanning to stop
	wg.Wait()
}

// Start begins the background scanning process
func (s *BackgroundScanner) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := s.scanValidatingAdmissionPolicies()
			if err != nil {
				klog.Errorf("Failed to scan ValidatingAdmissionPolicies: %v", err)
			}
			time.Sleep(5 * time.Minute) // Wait for 5 minutes before the next scan
		}
	}
}

// scanValidatingAdmissionPolicies scans ValidatingAdmissionPolicy resources in Kyverno
func (s *BackgroundScanner) scanValidatingAdmissionPolicies() error {
	listOptions := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/instance=kyverno",
	}

	validatingPolicies, err := s.kyvernoClient.KyvernoV1().ValidatingPolicies().List(context.Background(), listOptions)
	if err != nil {
		return fmt.Errorf("failed to list ValidatingPolicies: %v", err)
	}

	for _, policy := range validatingPolicies.Items {
		// Perform scanning logic here for each policy
		// ...

		// Record an event for the scanned policy
		s.eventRecorder.Event(&policy, v1.EventTypeNormal, "Scanned", "ValidatingAdmissionPolicy scanned successfully")
	}

	return nil
}
