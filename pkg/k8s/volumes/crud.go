package volumes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

const (
	maxRetries = 90
)

//Create deploys the volume claim for a given dev environment
func Create(ctx context.Context, dev *model.Dev, c *kubernetes.Clientset) error {
	vClient := c.CoreV1().PersistentVolumeClaims(dev.Namespace)
	pvc := translate()
	k8Volume, err := vClient.Get(pvc.Name, metav1.GetOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("error getting kubernetes volume claim: %s", err)
	}
	if k8Volume.Name == "" {
		log.Infof("creating volume claim '%s'...", pvc.Name)
		k8Volume, err = vClient.Create(pvc)
		if err != nil {
			return fmt.Errorf("error creating kubernetes volume claim: %s", err)
		}
	}

	tries := 0
	ticker := time.NewTicker(1 * time.Second)
	for k8Volume.Status.Phase != apiv1.ClaimBound {
		log.Debugf("PVC '%s' is '%s'...", k8Volume.Name, k8Volume.Status.Phase)
		select {
		case <-ticker.C:
			tries++
			if tries >= maxRetries {
				return fmt.Errorf("error creating PVC '%s'. It is '%s' after 150s", k8Volume.Name, k8Volume.Status.Phase)
			}
		case <-ctx.Done():
			log.Debug("cancelling call to create volume")
			return ctx.Err()
		}
		k8Volume, err = vClient.Get(pvc.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting kubernetes volume claim: %s", err)
		}
	}
	return nil
}

//Destroy destroys the volume claim for a given dev environment
func Destroy(name string, dev *model.Dev, c *kubernetes.Clientset) error {
	vClient := c.CoreV1().PersistentVolumeClaims(dev.Namespace)
	log.Infof("destroying volume claim '%s'...", name)

	ticker := time.NewTicker(1 * time.Second)
	for i := 0; i < maxRetries; i++ {
		err := vClient.Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				log.Infof("volume claim '%s' successfully destroyed", name)
				return nil
			}

			return fmt.Errorf("error getting kubernetes volume claim: %s", err)
		}

		<-ticker.C
		if i%10 == 5 {
			log.Infof("waiting for volume claim '%s' to be destroyed...", name)
		}
	}
	if err := checkIfAttached(name, dev, c); err != nil {
		return err
	}
	return fmt.Errorf("volume claim '%s' wasn't destroyed after 120s", name)
}

func checkIfAttached(pvc string, dev *model.Dev, c *kubernetes.Clientset) error {
	pods, err := c.CoreV1().Pods(dev.Namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Infof("failed to get available pods: %s", err)
		return nil
	}

	for _, p := range pods.Items {
		for _, v := range p.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				if v.PersistentVolumeClaim.ClaimName == pvc {
					log.Infof("pvc/%s is still attached to pod/%s", pvc, p.Name)
					return fmt.Errorf("can't delete your volume claim since it's still attached to 'pod/%s'", p.Name)
				}
			}
		}
	}

	return nil
}
