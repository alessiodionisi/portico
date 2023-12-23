package controller

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	ginformersv1 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1"
)

type Controller struct {
	context              context.Context
	gatewayClassInformer ginformersv1.GatewayClassInformer
	gatewayInformer      ginformersv1.GatewayInformer
	gatewayClientset     gclientset.Interface
	gatewayClassesSynced cache.InformerSynced
	objQueue             chan string
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	// c.gatewayClassInformer.Informer().Run(c.context.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), c.gatewayClassesSynced); !ok {
		return fmt.Errorf("error waiting for caches to sync")
	}

	slog.Info("Starting workers", "workers", workers)
	for i := 0; i < workers; i++ {
		go c.runWorker(ctx)
	}

	slog.Info("Started workers")
	<-ctx.Done()
	slog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker(ctx context.Context) {
	for {
		obj, more := <-c.objQueue
		if !more {
			return
		}

		_ = obj
	}
}

func (c *Controller) handleObject(obj interface{}) error {
	var object metav1.Object
	var objectType metav1.Type
	var ok bool

	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return fmt.Errorf("error decoding object, invalid type")
		}

		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return fmt.Errorf("error decoding object tombstone, invalid type")
		}

		slog.Debug("Recovered deleted object from tombstone", "name", object.GetName())
	}

	if objectType, ok = obj.(metav1.Type); !ok {
		return fmt.Errorf("error decoding object, invalid type")
	}

	slog.Debug("Handling object", "name", object.GetName(), "apiVersion", objectType.GetAPIVersion(), "kind", objectType.GetKind())

	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		slog.Debug("Object has owner", "name", object.GetName(), "owner", ownerRef.Name)
	}

	if objectType.GetAPIVersion() != "gateway.networking.k8s.io/v1" {
		slog.Debug("Object not handled", "name", object.GetName(), "apiVersion", objectType.GetAPIVersion(), "kind", objectType.GetKind())

		return nil
	}

	c.enqueueObject(obj)

	return nil
}

func (c *Controller) enqueueObject(obj interface{}) error {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return fmt.Errorf("error getting key for object: %w", err)
	}

	c.objQueue <- key

	return nil
}

// func (c *Controller) handleGatewayClass(gwc *gv1.GatewayClass) {
// 	// if gwc.Spec.ControllerName != "kubeportico.xyz/controller" {
// 	// 	slog.Debug("GatewayClass not handled", "name", gwc.Name, "controllerName", gwc.Spec.ControllerName)
// 	// 	return
// 	// }

// 	// gwcCopy := gwc.DeepCopy()

// 	// supportedVersion := "gateway.networking.k8s.io/v1"
// 	// if gwcCopy.APIVersion == supportedVersion {
// 	// 	gwcCopy.Status.Conditions = newDefaultGatewayClassConditions()
// 	// } else {
// 	// 	gwcCopy.Status.Conditions = newUnsupportedVersionGatewayClassConditions(supportedVersion)
// 	// }

// 	// if _, err := c.gatewayClientset.GatewayV1().GatewayClasses().UpdateStatus(c.context, gwcCopy, metav1.UpdateOptions{}); err != nil {
// 	// 	slog.Error("Error updating GatewayClass status", "name", gwc.Name, "error", err)
// 	// 	return
// 	// }
// }

func New(
	ctx context.Context,
	gatewayClientset gclientset.Interface,
	gatewayClassInformer ginformersv1.GatewayClassInformer,
	gatewayInformer ginformersv1.GatewayInformer,
) *Controller {
	c := &Controller{
		context:              ctx,
		gatewayClassInformer: gatewayClassInformer,
		gatewayClassesSynced: gatewayClassInformer.Informer().HasSynced,
		gatewayInformer:      gatewayInformer,
		gatewayClientset:     gatewayClientset,
		objQueue:             make(chan string, 300),
	}

	gatewayClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			gwc := obj.(*gatewayv1.GatewayClass)

			slog.Debug("GatewayClass added", "name", gwc.Name)

			if err := c.handleObject(obj); err != nil {
				slog.Error("Error handling GatewayClass", "name", gwc.Name, "error", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldGwc := oldObj.(*gatewayv1.GatewayClass)
			newGwc := newObj.(*gatewayv1.GatewayClass)

			if oldGwc.ResourceVersion == newGwc.ResourceVersion {
				return
			}

			slog.Debug("GatewayClass updated", "name", newGwc.Name)

			if err := c.handleObject(newObj); err != nil {
				slog.Error("Error handling GatewayClass", "name", newGwc.Name, "error", err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			gwc := obj.(*gatewayv1.GatewayClass)

			slog.Debug("GatewayClass deleted", "name", gwc.Name)

			if err := c.handleObject(obj); err != nil {
				slog.Error("Error handling GatewayClass", "name", gwc.Name, "error", err)
			}
		},
	})

	return c
}
