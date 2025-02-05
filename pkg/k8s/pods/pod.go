package pods

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/okteto/okteto/pkg/k8s/deployments"
	okLabels "github.com/okteto/okteto/pkg/k8s/labels"
	"github.com/okteto/okteto/pkg/k8s/replicasets"
	"github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	deploymentRevisionAnnotation = "deployment.kubernetes.io/revision"
	maxRetriesPodRunning         = 300 //1min pod is created
)

var (
	devTerminationGracePeriodSeconds int64
)

// GetDevPod returns the dev pod for a deployment
func GetDevPod(ctx context.Context, dev *model.Dev, c *kubernetes.Clientset, waitUntilDeployed bool) (*apiv1.Pod, error) {
	tries := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	for tries < maxRetriesPodRunning {
		pod, err := loopGetDevPod(ctx, dev, c, waitUntilDeployed)
		if err != nil {
			return nil, err
		}
		if pod != nil {
			return pod, nil
		}
		select {
		case <-ticker.C:
			tries++
			continue
		case <-ctx.Done():
			log.Debug("cancelling call to get dev pod")
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("kubernetes is taking too long to create the pod of your development environment. Please check for errors and try again")
}

func loopGetDevPod(ctx context.Context, dev *model.Dev, c *kubernetes.Clientset, waitUntilDeployed bool) (*apiv1.Pod, error) {
	d, err := deployments.GetRevisionAnnotatedDeploymentOrFailed(dev, c, waitUntilDeployed)
	if d == nil {
		return nil, err
	}

	log.Infof("deployment %s with revision %v is progressing", d.Name, d.Annotations[deploymentRevisionAnnotation])

	rs, err := replicasets.GetReplicaSetByDeployment(dev, d, c)
	if rs == nil {
		log.Infof("failed to get replicaset with revision %v: %s ", d.Annotations[deploymentRevisionAnnotation], err)
		return nil, err
	}

	log.Infof("replicaset %s with revison %s is progressing", rs.Name, d.Annotations[deploymentRevisionAnnotation])

	return getPodByReplicaSet(dev, rs, c)
}

func getPodByReplicaSet(dev *model.Dev, rs *appsv1.ReplicaSet, c *kubernetes.Clientset) (*apiv1.Pod, error) {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", okLabels.InteractiveDevLabel, dev.Name),
	}
	podList, err := c.CoreV1().Pods(dev.Namespace).List(opts)
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		for _, or := range pod.OwnerReferences {
			if or.UID == rs.UID {
				return &pod, nil
			}
		}
	}
	return nil, nil
}

//MonitorDevPod monitores the state of the pod
func MonitorDevPod(ctx context.Context, dev *model.Dev, pod *apiv1.Pod, c *kubernetes.Clientset, reporter chan string) (*apiv1.Pod, error) {
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", pod.Name),
	}

	watchPod, err := c.CoreV1().Pods(dev.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	opts = metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", pod.Name),
	}
	watchPodEvents, err := c.CoreV1().Events(dev.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case event := <-watchPod.ResultChan():
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Errorf("type error getting pod event")
				continue
			}
			log.Infof("pod %s updated", pod.Name)
			if pod.Status.Phase == apiv1.PodRunning {
				return pod, nil
			}
			if pod.DeletionTimestamp != nil {
				return nil, fmt.Errorf("development environment has been removed")
			}
		case event := <-watchPodEvents.ResultChan():
			e, ok := event.Object.(*v1.Event)
			if !ok {
				log.Errorf("type error getting pod event")
				continue
			}
			log.Infof("pod %s event: %s", pod.Name, e.Message)
			switch e.Reason {
			case "Failed", "FailedScheduling", "FailedCreatePodSandBox", "ErrImageNeverPull", "InspectFailed", "FailedCreatePodContainer":
				return nil, fmt.Errorf(e.Message)
			case "FailedAttachVolume", "FailedMount":
				reporter <- fmt.Sprintf("%s: retrying", e.Message)
			default:
				if e.Reason == "Pulling" {
					reporter <- strings.Replace(e.Message, "pulling", "Pulling", 1)
				}
			}
		case <-ctx.Done():
			log.Debug("cancelling call to monitor dev pod")
			return nil, ctx.Err()
		}
	}
}

//Exists returns if the dev pod still exists
func Exists(podName, namespace string, c *kubernetes.Clientset) bool {
	pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return pod.GetObjectMeta().GetDeletionTimestamp() == nil
}

// Restart restarts the pods of a deployment
func Restart(dev *model.Dev, c *kubernetes.Clientset) error {
	pods, err := c.CoreV1().Pods(dev.Namespace).List(
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", okLabels.DetachedDevLabel, dev.Name),
		},
	)
	if err != nil {
		log.Infof("error listing pods to restart: %s", err)
		return fmt.Errorf("failed to retrieve dev environment information")
	}

	for _, pod := range pods.Items {
		err := c.CoreV1().Pods(dev.Namespace).Delete(pod.Name, &metav1.DeleteOptions{GracePeriodSeconds: &devTerminationGracePeriodSeconds})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			return fmt.Errorf("error deleting kubernetes service: %s", err)
		}
	}

	return waitUntilRunning(dev.Namespace, fmt.Sprintf("%s=%s", okLabels.DetachedDevLabel, dev.Name), c)
}

func waitUntilRunning(namespace, selector string, c *kubernetes.Clientset) error {
	t := time.NewTicker(1 * time.Second)
	notready := map[string]bool{}

	for i := 0; i < 60; i++ {
		if i%5 == 0 {
			log.Infof("checking if pods are ready")
		}

		pods, err := c.CoreV1().Pods(namespace).List(
			metav1.ListOptions{
				LabelSelector: selector,
			},
		)

		if err != nil {
			log.Infof("error listing pods to check status after restart: %s", err)
			return fmt.Errorf("failed to retrieve dev environment information")
		}

		allRunning := true
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case apiv1.PodPending:
				allRunning = false
				notready[pod.GetName()] = true
			case apiv1.PodFailed:
				return fmt.Errorf("Pod %s failed to start", pod.Name)
			case apiv1.PodRunning:
				if isRunning(&pod) {
					if _, ok := notready[pod.GetName()]; ok {
						log.Infof("pod/%s is ready", pod.GetName())
						delete(notready, pod.GetName())
					}
				} else {
					allRunning = false
					notready[pod.GetName()] = true
					if i%5 == 0 {
						log.Infof("pod/%s is not ready", pod.GetName())
					}
				}
			}
		}

		if allRunning {
			log.Infof("pods are ready")
			return nil
		}

		<-t.C
	}

	pods := make([]string, 0, len(notready))
	for k := range notready {
		pods = append(pods, k)
	}

	return fmt.Errorf("Pod(s) %s didn't restart after 60 seconds", strings.Join(pods, ","))
}

func isRunning(p *apiv1.Pod) bool {
	if p.Status.Phase != apiv1.PodRunning {
		return false
	}

	if p.GetObjectMeta().GetDeletionTimestamp() != nil {
		return false
	}

	for _, c := range p.Status.Conditions {
		if c.Type == apiv1.PodReady {
			if c.Status == apiv1.ConditionTrue {
				return true
			}
		}
	}

	return false
}
