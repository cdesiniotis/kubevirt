/*
Copyright 2022 The KubeVirt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"

	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
	clonev1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/clone/v1alpha1"
	exportv1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/export/v1alpha1"
	instancetypev1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/instancetype/v1alpha1"
	migrationsv1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/migrations/v1alpha1"
	poolv1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/pool/v1alpha1"
	snapshotv1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/snapshot/v1alpha1"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	CloneV1alpha1() clonev1alpha1.CloneV1alpha1Interface
	ExportV1alpha1() exportv1alpha1.ExportV1alpha1Interface
	InstancetypeV1alpha1() instancetypev1alpha1.InstancetypeV1alpha1Interface
	MigrationsV1alpha1() migrationsv1alpha1.MigrationsV1alpha1Interface
	PoolV1alpha1() poolv1alpha1.PoolV1alpha1Interface
	SnapshotV1alpha1() snapshotv1alpha1.SnapshotV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	cloneV1alpha1        *clonev1alpha1.CloneV1alpha1Client
	exportV1alpha1       *exportv1alpha1.ExportV1alpha1Client
	instancetypeV1alpha1 *instancetypev1alpha1.InstancetypeV1alpha1Client
	migrationsV1alpha1   *migrationsv1alpha1.MigrationsV1alpha1Client
	poolV1alpha1         *poolv1alpha1.PoolV1alpha1Client
	snapshotV1alpha1     *snapshotv1alpha1.SnapshotV1alpha1Client
}

// CloneV1alpha1 retrieves the CloneV1alpha1Client
func (c *Clientset) CloneV1alpha1() clonev1alpha1.CloneV1alpha1Interface {
	return c.cloneV1alpha1
}

// ExportV1alpha1 retrieves the ExportV1alpha1Client
func (c *Clientset) ExportV1alpha1() exportv1alpha1.ExportV1alpha1Interface {
	return c.exportV1alpha1
}

// InstancetypeV1alpha1 retrieves the InstancetypeV1alpha1Client
func (c *Clientset) InstancetypeV1alpha1() instancetypev1alpha1.InstancetypeV1alpha1Interface {
	return c.instancetypeV1alpha1
}

// MigrationsV1alpha1 retrieves the MigrationsV1alpha1Client
func (c *Clientset) MigrationsV1alpha1() migrationsv1alpha1.MigrationsV1alpha1Interface {
	return c.migrationsV1alpha1
}

// PoolV1alpha1 retrieves the PoolV1alpha1Client
func (c *Clientset) PoolV1alpha1() poolv1alpha1.PoolV1alpha1Interface {
	return c.poolV1alpha1
}

// SnapshotV1alpha1 retrieves the SnapshotV1alpha1Client
func (c *Clientset) SnapshotV1alpha1() snapshotv1alpha1.SnapshotV1alpha1Interface {
	return c.snapshotV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.cloneV1alpha1, err = clonev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.exportV1alpha1, err = exportv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.instancetypeV1alpha1, err = instancetypev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.migrationsV1alpha1, err = migrationsv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.poolV1alpha1, err = poolv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.snapshotV1alpha1, err = snapshotv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.cloneV1alpha1 = clonev1alpha1.NewForConfigOrDie(c)
	cs.exportV1alpha1 = exportv1alpha1.NewForConfigOrDie(c)
	cs.instancetypeV1alpha1 = instancetypev1alpha1.NewForConfigOrDie(c)
	cs.migrationsV1alpha1 = migrationsv1alpha1.NewForConfigOrDie(c)
	cs.poolV1alpha1 = poolv1alpha1.NewForConfigOrDie(c)
	cs.snapshotV1alpha1 = snapshotv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.cloneV1alpha1 = clonev1alpha1.New(c)
	cs.exportV1alpha1 = exportv1alpha1.New(c)
	cs.instancetypeV1alpha1 = instancetypev1alpha1.New(c)
	cs.migrationsV1alpha1 = migrationsv1alpha1.New(c)
	cs.poolV1alpha1 = poolv1alpha1.New(c)
	cs.snapshotV1alpha1 = snapshotv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
