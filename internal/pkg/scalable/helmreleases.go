//nolint:dupl // this code is very similar for every resource, but its not really abstractable to avoid more duplication
package scalable

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
)

func getHelmReleaseKubeClient() client.Client {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = helmv2.AddToScheme(scheme)

	// init Kubernetes client
	kubeclient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	return kubeclient
}

// getHelmReleases is the getResourceFunc for HelmReleases.
func getHelmReleases(namespace string, clientsets *Clientsets, ctx context.Context) ([]Workload, error) {
	kubeclient := getHelmReleaseKubeClient()
	namespacedkubeclient := client.NewNamespacedClient(kubeclient, namespace)

	helmreleases := &helmv2.HelmReleaseList{}

	err := namespacedkubeclient.List(ctx, helmreleases)
	if err != nil {
		return nil, fmt.Errorf("failed to get helmreleases: %w", err)
	}

	results := make([]Workload, 0, len(helmreleases.Items))
	for i := range helmreleases.Items {
		results = append(results, &suspendScaledWorkload{&helmrelease{&helmreleases.Items[i]}})
	}

	return results, nil
}

// helmrelease is a wrapper for helm/v2.HelmRelease to implement the suspendScaledResource interface.
type helmrelease struct {
	*helmv2.HelmRelease
}

// setSuspend sets the value of the suspend field on the helmrelease.
func (h *helmrelease) setSuspend(suspend bool) {
	h.Spec.Suspend = suspend
}

// Update updates the resource with all changes made to it. It should only be called once on a resource.
func (h *helmrelease) Update(clientsets *Clientsets, ctx context.Context) error {
	kubeclient := getHelmReleaseKubeClient()

	err := kubeclient.Update(ctx, h)
	if err != nil {
		return fmt.Errorf("failed to update helmrelease: %w", err)
	}

	return nil
}
