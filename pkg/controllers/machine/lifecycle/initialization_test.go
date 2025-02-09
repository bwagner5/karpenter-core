/*
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

package lifecycle_test

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/aws/karpenter-core/pkg/cloudprovider/fake"
	"github.com/aws/karpenter-core/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/aws/karpenter-core/pkg/test/expectations"
)

var _ = Describe("Initialization", func() {
	var provisioner *v1alpha5.Provisioner

	BeforeEach(func() {
		provisioner = test.Provisioner()
	})
	It("should consider the Machine initialized when all initialization conditions are met", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("50Mi"),
						v1.ResourcePods:   resource.MustParse("5"),
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
		})
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))

		node = ExpectExists(ctx, env.Client, node)
		node.Status.Capacity = v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("10"),
			v1.ResourceMemory: resource.MustParse("100Mi"),
			v1.ResourcePods:   resource.MustParse("110"),
		}
		node.Status.Allocatable = v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("8"),
			v1.ResourceMemory: resource.MustParse("80Mi"),
			v1.ResourcePods:   resource.MustParse("110"),
		}
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionTrue))
	})
	It("should add the initialization label to the node when the Machine is initialized", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("50Mi"),
						v1.ResourcePods:   resource.MustParse("5"),
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		})
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		node = ExpectExists(ctx, env.Client, node)
		Expect(node.Labels).To(HaveKeyWithValue(v1alpha5.LabelNodeInitialized, "true"))
	})
	It("should not consider the Node to be initialized when the status of the Node is NotReady", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("50Mi"),
						v1.ResourcePods:   resource.MustParse("5"),
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			ReadyStatus: v1.ConditionFalse,
		})
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))
	})
	It("should not consider the Node to be initialized when all requested resources aren't registered", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:          resource.MustParse("2"),
						v1.ResourceMemory:       resource.MustParse("50Mi"),
						v1.ResourcePods:         resource.MustParse("5"),
						fake.ResourceGPUVendorA: resource.MustParse("1"),
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		// Update the machine to add mock the instance type having an extended resource
		machine.Status.Capacity[fake.ResourceGPUVendorA] = resource.MustParse("2")
		machine.Status.Allocatable[fake.ResourceGPUVendorA] = resource.MustParse("2")
		ExpectApplied(ctx, env.Client, machine)

		// Extended resource hasn't registered yet by the daemonset
		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		})
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))
	})
	It("should consider the node to be initialized once all the resources are registered", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:          resource.MustParse("2"),
						v1.ResourceMemory:       resource.MustParse("50Mi"),
						v1.ResourcePods:         resource.MustParse("5"),
						fake.ResourceGPUVendorA: resource.MustParse("1"),
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		// Update the machine to add mock the instance type having an extended resource
		machine.Status.Capacity[fake.ResourceGPUVendorA] = resource.MustParse("2")
		machine.Status.Allocatable[fake.ResourceGPUVendorA] = resource.MustParse("2")
		ExpectApplied(ctx, env.Client, machine)

		// Extended resource hasn't registered yet by the daemonset
		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		})
		ExpectApplied(ctx, env.Client, node)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))

		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))

		// Node now registers the resource
		node = ExpectExists(ctx, env.Client, node)
		node.Status.Capacity[fake.ResourceGPUVendorA] = resource.MustParse("2")
		node.Status.Allocatable[fake.ResourceGPUVendorA] = resource.MustParse("2")
		ExpectApplied(ctx, env.Client, node)

		// Reconcile the machine and the Machine/Node should now be initilized
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionTrue))
	})
	It("should not consider the Node to be initialized when all startupTaints aren't removed", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("50Mi"),
						v1.ResourcePods:   resource.MustParse("5"),
					},
				},
				StartupTaints: []v1.Taint{
					{
						Key:    "custom-startup-taint",
						Effect: v1.TaintEffectNoSchedule,
						Value:  "custom-startup-value",
					},
					{
						Key:    "other-custom-startup-taint",
						Effect: v1.TaintEffectNoExecute,
						Value:  "other-custom-startup-value",
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		})
		ExpectApplied(ctx, env.Client, node)

		// Should add the startup taints to the node
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		node = ExpectExists(ctx, env.Client, node)
		Expect(node.Spec.Taints).To(ContainElements(
			v1.Taint{
				Key:    "custom-startup-taint",
				Effect: v1.TaintEffectNoSchedule,
				Value:  "custom-startup-value",
			},
			v1.Taint{
				Key:    "other-custom-startup-taint",
				Effect: v1.TaintEffectNoExecute,
				Value:  "other-custom-startup-value",
			},
		))

		// Shouldn't consider the node ready since the startup taints still exist
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))
	})
	It("should consider the Node to be initialized once the startup taints are removed", func() {
		machine := test.Machine(v1alpha5.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1alpha5.ProvisionerNameLabelKey: provisioner.Name,
				},
			},
			Spec: v1alpha5.MachineSpec{
				Resources: v1alpha5.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("50Mi"),
						v1.ResourcePods:   resource.MustParse("5"),
					},
				},
				StartupTaints: []v1.Taint{
					{
						Key:    "custom-startup-taint",
						Effect: v1.TaintEffectNoSchedule,
						Value:  "custom-startup-value",
					},
					{
						Key:    "other-custom-startup-taint",
						Effect: v1.TaintEffectNoExecute,
						Value:  "other-custom-startup-value",
					},
				},
			},
		})
		ExpectApplied(ctx, env.Client, provisioner, machine)
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)

		node := test.Node(test.NodeOptions{
			ProviderID: machine.Status.ProviderID,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("100Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("8"),
				v1.ResourceMemory: resource.MustParse("80Mi"),
				v1.ResourcePods:   resource.MustParse("110"),
			},
		})
		ExpectApplied(ctx, env.Client, node)

		// Shouldn't consider the node ready since the startup taints still exist
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionFalse))

		node = ExpectExists(ctx, env.Client, node)
		node.Spec.Taints = []v1.Taint{}
		ExpectApplied(ctx, env.Client, node)

		// Machine should now be ready since all startup taints are removed
		ExpectReconcileSucceeded(ctx, machineController, client.ObjectKeyFromObject(machine))
		machine = ExpectExists(ctx, env.Client, machine)
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineRegistered).Status).To(Equal(v1.ConditionTrue))
		Expect(ExpectStatusConditionExists(machine, v1alpha5.MachineInitialized).Status).To(Equal(v1.ConditionTrue))
	})
})
