package cluster

import (
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	api "github.com/xoe-labs/odoo-operator/pkg/apis/odoo/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func builder(into runtime.Object, c *api.OdooCluster, i ...int) (string, error) {
	syncer(into, c, i...)
	switch o := into.(type) {

	case *api.PgNamespace:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	case *v1.PersistentVolumeClaim:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	case *v1.ConfigMap:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	case *appsv1.Deployment:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	case *v1.Service:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	case *v1.Secret:
		addOwnerRefToObject(o, asOwner(c))
		return o.GetName(), nil

	}

	return "", nil
}

func syncer(into runtime.Object, c *api.OdooCluster, i ...int) (bool, error) {
	changed := false
	switch o := into.(type) {

	case *api.PgNamespace:
		newSpec := c.Spec.PgSpec
		if !reflect.DeepEqual(o.Spec, newSpec) {
			changed = true
			o.Spec = newSpec
		}
		logrus.Debugf("Syncer (PgNamespace-Spec) +++++ %+v", o.Spec)
		return changed, nil

	case *v1.PersistentVolumeClaim:
		newSpec := c.Spec.Volumes[i[0]].Spec
		if !reflect.DeepEqual(o.Spec, newSpec) {
			changed = true
			o.Spec = newSpec
		}
		logrus.Debugf("Syncer (PVC-Spec) +++++ %+v", o.Spec)
		return changed, nil

	case *v1.ConfigMap:
		var cfgDefaultData string
		var cfgCustomData string

		cfgDefaultData = newConfigWithDefaultParams(cfgDefaultData)
		newSpec := map[string]string{odooDefaultConfig: cfgDefaultData}
		if len(c.Spec.ConfigMap) != 0 {
			cfgCustomData = c.Spec.ConfigMap
			newSpec[odooCustomConfig] = cfgCustomData
		}
		if !reflect.DeepEqual(o.Data, newSpec) {
			changed = true
			o.Data = newSpec
		}
		logrus.Debugf("Syncer (ConfigMap-Spec) +++++ %+v", o.Data)
		return changed, nil

	case *appsv1.Deployment:
		volumes := []v1.Volume{
			{
				Name: configVolName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: c.GetName(),
						},
						DefaultMode: func(a int32) *int32 { return &a }(420),
					},
				},
			},
		}

		for _, s := range c.Spec.Volumes {
			vol := v1.Volume{
				// kubernetes.io/pvc-protection
				Name: volumeNameForOdoo(c, &s),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: volumeNameForOdoo(c, &s),
						ReadOnly:  false,
					},
				},
			}
			volumes = append(volumes, vol)

		}

		if !reflect.DeepEqual(o.Spec.Template.Spec.Volumes, volumes) {
			changed = true
			o.Spec.Template.Spec.Volumes = volumes
		}

		securityContext := &v1.PodSecurityContext{
			RunAsUser:    func(i int64) *int64 { return &i }(9001),
			RunAsNonRoot: func(b bool) *bool { return &b }(true),
			FSGroup:      func(i int64) *int64 { return &i }(9001),
		}

		if !reflect.DeepEqual(o.Spec.Template.Spec.SecurityContext, securityContext) {
			changed = true
			o.Spec.Template.Spec.SecurityContext = securityContext
		}

		trackSpec := c.Spec.Tracks[i[0]]
		tierSpec := c.Spec.Tiers[i[1]]
		newContainers := []v1.Container{odooContainer(c, &trackSpec, &tierSpec)}

		if !reflect.DeepEqual(o.Spec.Template.Spec.Containers, newContainers) {
			// logrus.Errorf("OldContainers %+v", o.Spec.Template.Spec.Containers)
			// logrus.Errorf("NewContainers %+v", newContainers)
			// logrus.Error("NewContainers")
			changed = true
			o.Spec.Template.Spec.Containers = newContainers
		}
		o.Spec.Template.ObjectMeta = o.ObjectMeta

		selector := selectorForOdooCluster(c.GetName())

		if !reflect.DeepEqual(o.Spec.Selector, &metav1.LabelSelector{MatchLabels: selector}) {
			changed = true
			o.Spec.Selector = &metav1.LabelSelector{MatchLabels: selector}
		}
		if !reflect.DeepEqual(o.Spec.Replicas, &tierSpec.Replicas) {
			changed = true
			o.Spec.Replicas = &tierSpec.Replicas
		}

		newStrategy := appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
			},
		}
		if !reflect.DeepEqual(o.Spec.Strategy, newStrategy) {
			changed = true
			o.Spec.Strategy = newStrategy
		}
		logrus.Debugf("Syncer (Deployment-Spec) +++++ %+v", o.Spec)
		return changed, nil

	case *v1.Service:
		selector := selectorForOdooCluster(c.GetName())
		var svcPorts []v1.ServicePort

		tierSpec := c.Spec.Tiers[i[1]]

		switch tierSpec.Name {
		case api.ServerTier:
			svcPorts = []v1.ServicePort{{
				Name:       clientPortName,
				Protocol:   v1.ProtocolTCP,
				Port:       int32(clientPort),
				TargetPort: intstr.FromString(clientPortName),
			}}
		case api.LongpollingTier:
			svcPorts = []v1.ServicePort{{
				Name:       longpollingPortName,
				Protocol:   v1.ProtocolTCP,
				Port:       int32(longpollingPort),
				TargetPort: intstr.FromString(longpollingPortName),
			}}
		}

		if !reflect.DeepEqual(o.Spec.Selector, selector) {
			changed = true
			o.Spec.Selector = selector
		}
		if !reflect.DeepEqual(o.Spec.Ports, svcPorts) {
			changed = true
			o.Spec.Ports = svcPorts
		}
		logrus.Debugf("Syncer (Service-Spec) +++++ %+v", o.Spec)
		return changed, nil

	}
	return changed, nil
}

// volumeNameForOdoo is the volume name for the given odoo cluster.
func volumeNameForOdoo(cr *api.OdooCluster, s *api.Volume) string {
	return cr.GetName() + strings.ToLower(string(s.Name))
}
