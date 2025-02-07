package main

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
)

func main() {
	namespacedKubeClient := getKubeClient("flux-system")
	// set a deadline for the Kubernetes API operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	helmReleaseList := listHelmReleases(namespacedKubeClient, ctx)
	firstHelmRelease := helmReleaseList.Items[0]
	suspendHelmRelease(namespacedKubeClient, firstHelmRelease, ctx)
}

func suspendHelmRelease(namespacedKubeClient client.Client, helmRelease helmv2.HelmRelease, ctx context.Context) {
	helmRelease.Spec.Suspend = true
	namespacedKubeClient.Update(ctx, &helmRelease)
}

func listHelmReleases(namespacedKubeClient client.Client, ctx context.Context) helmv2.HelmReleaseList {
	helmReleaseList := &helmv2.HelmReleaseList{}
	if err := namespacedKubeClient.List(ctx, helmReleaseList); err != nil {
		fmt.Println(err)
	}
	return *helmReleaseList
}

func getKubeClient(ns string) client.Client {
	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = helmv2.AddToScheme(scheme)

	// init Kubernetes client
	kubeClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}
	namespacedKubeClient := client.NewNamespacedClient(kubeClient, ns)

	return namespacedKubeClient
}
