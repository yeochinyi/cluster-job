package clusterjob

import (
	"context"
	appv1alpha1 "github.com/yeochinyi/cluster-job/pkg/apis/app/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clusterjob")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ClusterJob Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterJob{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterjob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterJob
	err = c.Watch(&source.Kind{Type: &appv1alpha1.ClusterJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ClusterJob
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.ClusterJob{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileClusterJob implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClusterJob{}

// ReconcileClusterJob reconciles a ClusterJob object
type ReconcileClusterJob struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ClusterJob object and makes changes based on the state read
// and what is in the ClusterJob.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ClusterJob")

	// Fetch the ClusterJob instance
	instance := &appv1alpha1.ClusterJob{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	jobs := newJobsForCR(instance)

	instance.Status.TotalStarted = uint8(len(jobs))
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Failed to update status")
		return reconcile.Result{}, err
	}

	// Set ClusterJob instance as the owner and controller
	for _, job := range jobs {
		if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		found := &batchv1.Job{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Job", "Pod.Namespace", job.Namespace, "Pod.Name", job.Name)
			err = r.client.Create(context.TODO(), job)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			// Real Error
			return reconcile.Result{}, err
		}

		var status appv1alpha1.JobStatus

		switch {
		case found.Status.Failed > 0:
			status = appv1alpha1.FAILED
		case found.Status.Succeeded > 0:
			status = appv1alpha1.SUCCEEDED
		case found.Status.Active > 0:
			status = appv1alpha1.ACTIVE

		}

		instance.Status.JobStatuses[job.Name] = status

		//found.Status.Failed
		//found.Status.Active

		//// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Job already exists", "Job.Namespace", found.Namespace, "Job.Name", found.Name)
	}

	return reconcile.Result{}, nil
}

// newJobForCR returns a busybox pod with the same name/namespace as the cr
func newJobsForCR(cr *appv1alpha1.ClusterJob) []*batchv1.Job {
	labels := map[string]string{
		"app": cr.Name,
	}

	var jobs []*batchv1.Job

	for key, value := range cr.Spec.JobImages {

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name + "-" + key,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: batchv1.JobSpec{
				Parallelism:           nil,
				Completions:           nil,
				ActiveDeadlineSeconds: nil,
				BackoffLimit:          nil,
				Selector:              nil,
				ManualSelector:        nil,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.PodSpec{
						Volumes:        nil,
						InitContainers: nil,
						Containers: []corev1.Container{
							{
								Name:  key,
								Image: value,
								//Command: []string{"sleep", "3600"},
							},
						},
						RestartPolicy:                 corev1.RestartPolicyNever,
						TerminationGracePeriodSeconds: nil,
						ActiveDeadlineSeconds:         nil,
						DNSPolicy:                     "",
						NodeSelector:                  nil,
						ServiceAccountName:            "",
						AutomountServiceAccountToken:  nil,
						NodeName:                      "",
						HostNetwork:                   false,
						HostPID:                       false,
						HostIPC:                       false,
						ShareProcessNamespace:         nil,
						SecurityContext:               nil,
						ImagePullSecrets:              nil,
						Hostname:                      "",
						Subdomain:                     "",
						Affinity:                      nil,
						SchedulerName:                 "",
						Tolerations:                   nil,
						HostAliases:                   nil,
						PriorityClassName:             "",
						Priority:                      nil,
						DNSConfig:                     nil,
						ReadinessGates:                nil,
						RuntimeClassName:              nil,
						EnableServiceLinks:            nil,
						PreemptionPolicy:              nil,
					},
				},
				TTLSecondsAfterFinished: nil,
			},
		}

		jobs = append(jobs, job)
	}

	return jobs

}
