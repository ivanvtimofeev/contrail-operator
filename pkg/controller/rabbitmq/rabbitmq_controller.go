package rabbitmq

import (
	"context"

	"time"

	"github.com/Juniper/contrail-operator/pkg/apis/contrail/v1alpha1"
	"github.com/Juniper/contrail-operator/pkg/controller/utils"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mRand "math/rand"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_rabbitmq")

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = mRand.NewSource(time.Now().UnixNano())

func resourceHandler(myclient client.Client) handler.Funcs {
	appHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.RabbitmqList{}
			err := myclient.List(context.TODO(), list, listOps)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.MetaNew.GetNamespace()}
			list := &v1alpha1.RabbitmqList{}
			err := myclient.List(context.TODO(), list, listOps)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.MetaNew.GetNamespace(),
					}})
				}
			}
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.RabbitmqList{}
			err := myclient.List(context.TODO(), list, listOps)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
		GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.RabbitmqList{}
			err := myclient.List(context.TODO(), list, listOps)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
	}
	return appHandler
}

// Add adds the Rabbitmq controller to the manager.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitmq{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Manager: mgr}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller.
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rabbitmq.
	if err = c.Watch(&source.Kind{Type: &v1alpha1.Rabbitmq{}},
		&handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	serviceMap := map[string]string{"contrail_manager": "rabbitmq"}
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := resourceHandler(mgr.GetClient())
	predInitStatus := utils.PodInitStatusChange(serviceMap)
	predPodIPChange := utils.PodIPChange(serviceMap)
	predInitRunning := utils.PodInitRunning(serviceMap)

	if err = c.Watch(srcPod, podHandler, predPodIPChange); err != nil {
		return err
	}
	if err = c.Watch(srcPod, podHandler, predInitStatus); err != nil {
		return err
	}
	if err = c.Watch(srcPod, podHandler, predInitRunning); err != nil {
		return err
	}

	srcManager := &source.Kind{Type: &v1alpha1.Manager{}}
	managerHandler := resourceHandler(mgr.GetClient())
	predManagerSizeChange := utils.ManagerSizeChange(utils.RabbitmqGroupKind())
	if err = c.Watch(srcManager, managerHandler, predManagerSizeChange); err != nil {
		return err
	}

	srcSTS := &source.Kind{Type: &appsv1.StatefulSet{}}
	stsHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Rabbitmq{},
	}
	stsPred := utils.STSStatusChange(utils.RabbitmqGroupKind())
	if err = c.Watch(srcSTS, stsHandler, stsPred); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRabbitmq implements reconcile.Reconciler.
var _ reconcile.Reconciler = &ReconcileRabbitmq{}

// ReconcileRabbitmq reconciles a Rabbitmq object.
type ReconcileRabbitmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver.
	Client  client.Client
	Scheme  *runtime.Scheme
	Manager manager.Manager
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func randomString(size int) string {
	b := make([]byte, size)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := size-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// Reconcile reconciles the Rabbitmq resource.
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq")
	instanceType := "rabbitmq"
	instance := &v1alpha1.Rabbitmq{}
	if err := r.Client.Get(context.TODO(), request.NamespacedName, instance); err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	managerInstance, err := instance.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Rabbitmq != nil {
			rabbitmqManagerInstance := managerInstance.Spec.Services.Rabbitmq
			instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
				managerInstance.Spec.CommonConfiguration,
				rabbitmqManagerInstance.Spec.CommonConfiguration)
			if err = r.Client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	configMap, err := instance.CreateConfigMap(request.Name+"-"+instanceType+"-configmap", r.Client, r.Scheme, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	configMap2, err := instance.CreateConfigMap(request.Name+"-"+instanceType+"-configmap-runner", r.Client, r.Scheme, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	secret, err := instance.CreateSecret(request.Name+"-secret", r.Client, r.Scheme, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	secretCertificates, err := instance.CreateSecret(request.Name+"-secret-certificates", r.Client, r.Scheme, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	statefulSet := GetSTS()
	if err = instance.PrepareSTS(statefulSet, &instance.Spec.CommonConfiguration, request, r.Scheme, r.Client); err != nil {
		return reconcile.Result{}, err
	}

	instance.AddVolumesToIntendedSTS(statefulSet,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume",
			configMap2.Name: request.Name + "-" + instanceType + "-runner"})
	instance.AddSecretVolumesToIntendedSTS(statefulSet, map[string]string{secretCertificates.Name: request.Name + "-secret-certificates"})

	for idx, container := range statefulSet.Spec.Template.Spec.Containers {

		if container.Name == "rabbitmq" {
			command := []string{"bash", "/runner/run.sh"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			if instance.Spec.ServiceConfiguration.Containers[container.Name].Command == nil {
				(&statefulSet.Spec.Template.Spec.Containers[idx]).Command = command
			} else {
				(&statefulSet.Spec.Template.Spec.Containers[idx]).Command = instance.Spec.ServiceConfiguration.Containers[container.Name].Command
			}
			volumeMountList := []corev1.VolumeMount{}

			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/rabbitmq",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			volumeMount = corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-runner",
				MountPath: "/runner/",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			volumeMount = corev1.VolumeMount{
				Name:      request.Name + "-secret-certificates",
				MountPath: "/etc/certificates",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&statefulSet.Spec.Template.Spec.Containers[idx]).EnvFrom = []corev1.EnvFromSource{{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: request.Name + "-" + instanceType + "-configmap",
					},
				},
			}}
			(&statefulSet.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&statefulSet.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Containers[container.Name].Image
		}
	}

	// Configure InitContainers.
	for idx, container := range statefulSet.Spec.Template.Spec.InitContainers {
		(&statefulSet.Spec.Template.Spec.InitContainers[idx]).Image = instance.Spec.ServiceConfiguration.Containers[container.Name].Image
		if instance.Spec.ServiceConfiguration.Containers[container.Name].Command != nil {
			(&statefulSet.Spec.Template.Spec.InitContainers[idx]).Command = instance.Spec.ServiceConfiguration.Containers[container.Name].Command
		}
	}

	//err = r.Client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: instance.Namespace}, secret)
	//if err != nil && errors.IsNotFound(err) {
	//	return reconcile.Result{}, err
	//}

	var password string
	var user string
	var vhost string
	if instance.Spec.ServiceConfiguration.Password != "" {
		password = instance.Spec.ServiceConfiguration.Password
	} else {
		password = randomString(32)
	}

	if instance.Spec.ServiceConfiguration.User != "" {
		user = instance.Spec.ServiceConfiguration.User
	} else {
		user = randomString(8)
	}
	if instance.Spec.ServiceConfiguration.Vhost != "" {
		vhost = instance.Spec.ServiceConfiguration.Vhost
	} else {
		vhost = randomString(6)
	}

	secretPassword := []byte(password)
	secretUser := []byte(user)
	secretVhost := []byte(vhost)
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	if _, ok := secret.Data["password"]; !ok {
		secret.Data["password"] = secretPassword
	}
	if _, ok := secret.Data["user"]; !ok {
		secret.Data["user"] = secretUser
	}
	if _, ok := secret.Data["vhost"]; !ok {
		secret.Data["vhost"] = secretVhost
	}

	if err := r.Client.Update(context.Background(), secret); err != nil {
		return reconcile.Result{}, err
	}

	if err = instance.CreateSTS(statefulSet, &instance.Spec.CommonConfiguration, instanceType, request, r.Scheme, r.Client); err != nil {
		return reconcile.Result{}, err
	}

	if err = instance.UpdateSTS(statefulSet, &instance.Spec.CommonConfiguration, instanceType, request, r.Scheme, r.Client, "rolling"); err != nil {
		return reconcile.Result{}, err
	}

	podIPList, podIPMap, err := instance.PodIPListAndIPMapFromInstance(instanceType, request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		if err = instance.InstanceConfiguration(request, podIPList, r.Client); err != nil {
			return reconcile.Result{}, err
		}
		hostNetwork := true
		if instance.Spec.CommonConfiguration.HostNetwork != nil {
			hostNetwork = *instance.Spec.CommonConfiguration.HostNetwork
		}
		if err = v1alpha1.CreateAndSignCsr(r.Client, request, r.Scheme, instance, r.Manager.GetConfig(), podIPList, hostNetwork); err != nil {
			return reconcile.Result{}, err
		}
		if err = instance.SetPodsToReady(podIPList, r.Client); err != nil {
			return reconcile.Result{}, err
		}

		if err = instance.ManageNodeStatus(podIPMap, r.Client); err != nil {
			return reconcile.Result{}, err
		}
	}
	if instance.Status.Active == nil {
		active := false
		instance.Status.Active = &active
	}

	if err = instance.SetInstanceActive(r.Client, instance.Status.Active, statefulSet, request); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
