package sopactions

import (
	"context"
	logger "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func UpgradeRHSSO(ctx context.Context, client client.Client) {
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
		logger.Error("Error upgrading RHSSO")
	}
}
