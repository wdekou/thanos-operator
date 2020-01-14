package resources

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	"github.com/imdario/mergo"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	nameLabel      = "app.kubernetes.io/name"
	instanceLabel  = "app.kubernetes.io/instance"
	versionLabel   = "app.kubernetes.io/version"
	componentLabel = "app.kubernetes.io/component"
	managedByLabel = "app.kubernetes.io/managed-by"
)

type ThanosComponentReconciler struct {
	Thanos      *v1alpha1.Thanos
	ObjectSores []v1alpha1.ObjectStore
	*reconciler.GenericResourceReconciler
}

func (t *ThanosComponentReconciler) Reconcile() (*reconcile.Result, error) {
	resourceList := []Resource{
		t.queryDeployment,
		t.storeDeployment,
		t.storeService,
	}
	// Generate objects from resources
	for _, res := range resourceList {
		o, state, err := res()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to create desired object")
		}
		if o == nil {
			return nil, errors.Errorf("Reconcile error! Resource %#v returns with nil object", res)
		}
		result, err := t.ReconcileResource(o, state)
		if err != nil {
			return nil, errors.WrapIf(err, "failed to reconcile resource")
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, nil
}

func (t *ThanosComponentReconciler) setDefaults() error {
	if t.Thanos.Spec.Query != nil {
		err := mergo.Merge(t.Thanos.Spec.Query, v1alpha1.DefaultQuery)
		if err != nil {
			return err
		}
	}
	if t.Thanos.Spec.StoreGateway != nil {
		err := mergo.Merge(t.Thanos.Spec.StoreGateway, v1alpha1.DefaultStoreGateway)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewThanosComponentReconciler(thanos *v1alpha1.Thanos, objectStores *v1alpha1.ObjectStoreList, genericReconciler *reconciler.GenericResourceReconciler) (*ThanosComponentReconciler, error) {
	reconciler := &ThanosComponentReconciler{
		Thanos:                    thanos,
		ObjectSores:               objectStores.Items,
		GenericResourceReconciler: genericReconciler,
	}
	err := reconciler.setDefaults()
	return reconciler, err
}
