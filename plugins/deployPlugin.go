package plugins

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
)

// DeploymentReconciler reconciles a Bread object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (d *DeploymentReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	deployObject := &appv1.Deployment{}

	if err := d.Client.Get(ctx, request.NamespacedName, deployObject); err != nil {
		return reconcile.Result{}, err
	}

	// 获取所有的 annotations
	annos := deployObject.ObjectMeta.Annotations
	fmt.Println("-------------begin get annos--------------")
	for anno := range annos {
		fmt.Println("key: " + anno)
		fmt.Println("values: " + annos[anno])
	}
	fmt.Println("-------------finish get annos--------------")

	//获取 deployment 对应的 pod
	selector := labels.NewSelector()
	for k, v := range deployObject.Spec.Selector.MatchLabels {
		r, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			klog.Error(err)
			continue
		}
		req := labels.Requirement{}
		r.DeepCopyInto(&req)
		selector = selector.Add(req)
	}
	pods := &corev1.PodList{}
	d.Client.List(ctx, pods, &client.ListOptions{LabelSelector: selector})

	// update pod
	for i, pod := range pods.Items {
		tmp := pod.DeepCopy()
		patch := client.MergeFrom(&pod)

		//修改你想要的字段，我这边改的是Annotations
		if tmp.Annotations == nil {
			tmp.Annotations = make(map[string]string)
			tmp.Annotations[tmp.Name] = "test" + strconv.Itoa(i)
			err := d.Client.Patch(ctx, tmp, patch)
			if err != nil {
				fmt.Println("update err: " + err.Error())
			}
		}
	}

	return reconcile.Result{}, nil
}

func (d *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.Deployment{}).
		Complete(d)
}