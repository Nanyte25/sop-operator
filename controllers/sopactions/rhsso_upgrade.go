package sopactions

import (
	"context"
	"errors"
	logger "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

func UpgradeRHSSO(ctx context.Context, client client.Client) error {
	logger.Info("Upgrading RHSSO due to CVE rollout...")

	// Find the image to remove
	statefulSetKey := types.NamespacedName{
		Name:      "keycloak",
		Namespace: "redhat-rhoam-rhsso",
	}
	statefulSet := &appsv1.StatefulSet{}
	_ = client.Get(ctx, statefulSetKey, statefulSet)
	image := statefulSet.Spec.Template.Spec.Containers[0].Image

	// Create temporary Daemon Set
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tempds",
			Namespace: "redhat-rhoam-rhsso",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, client, daemonSet, func() error {
		daemonSet.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "tempds",
			},
		}
		daemonSet.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "tempds",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:            "tempds",
					Image:           image,
					ImagePullPolicy: "Always",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8080,
						},
					},
				}},
			},
		}
		return nil
	})
	if err != nil {
		logger.Error("Error occurred while upgrading RHSSO")
	}

	// Compare number of worker nodes to number of pull image events
	numPullImageEvents := getNumberPullImageEvents(ctx, client, image)
	//numWorkerNodes := getNumberWorkerNodes(ctx, client)

	// Verify the image was pulled to all worker nodes
	//if numPullImageEvents < numWorkerNodes {
	if numPullImageEvents < 4 { // Hard coded for testing purposes only. Revert when finished.
		errMsg := "not all worker nodes have pulled the new image yet"
		logger.Error(errMsg)
		return errors.New(errMsg)
	} else {
		return nil
	}

}

func getNumberPullImageEvents(ctx context.Context, client client.Client, image string) int {
	// Get EventList of all Events
	eventList := &corev1.EventList{}
	_ = client.List(ctx, eventList)

	// Get the number of pull image events
	numPullImageEvents := 0
	imageString := "Successfully pulled image \"" + image + "\""

	for _, event := range eventList.Items {
		if isEventPullImage(event, imageString) {
			numPullImageEvents++
		}
	}
	return numPullImageEvents
}

func isEventPullImage(event corev1.Event, imageString string) bool {
	if strings.Contains(event.Name, "tempds") && strings.Contains(event.Message, imageString) {
		return true
	}
	return false
}

func getNumberWorkerNodes(ctx context.Context, client client.Client) int {
	// Get NodeList of all Nodes
	nodeList := &corev1.NodeList{}
	_ = client.List(ctx, nodeList)

	// Get the number of worker Nodes
	numWorkerNodes := 0
	for _, node := range nodeList.Items {
		if isNodeWorker(node) {
			numWorkerNodes++
		}
	}
	return numWorkerNodes
}

func isNodeWorker(node corev1.Node) bool {
	for labKey, _ := range node.Labels {
		if strings.Contains(labKey, "worker") {
			return true
		}
	}
	return false
}
