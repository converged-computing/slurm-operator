/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"fmt"

	api "github.com/converged-computing/slurm-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// getResourceGroup can return a ResourceList for either requests or limits
func (r *SlurmReconciler) getResourceGroup(
	items api.Resource,
) (corev1.ResourceList, error) {

	r.Log.Info("🍅️ Resource", "items", items)
	list := corev1.ResourceList{}
	for key, unknownValue := range items {
		if unknownValue.Type == intstr.Int {

			value := unknownValue.IntVal
			r.Log.Info("🍅️ ResourceKey", "Key", key, "Value", value)
			limit, err := resource.ParseQuantity(fmt.Sprintf("%d", value))
			if err != nil {
				return list, err
			}

			if key == "memory" {
				list[corev1.ResourceMemory] = limit
			} else if key == "cpu" {
				list[corev1.ResourceCPU] = limit
			} else {
				list[corev1.ResourceName(key)] = limit
			}
		} else if unknownValue.Type == intstr.String {

			value := unknownValue.StrVal
			r.Log.Info("🍅️ ResourceKey", "Key", key, "Value", value)
			if key == "memory" {
				list[corev1.ResourceMemory] = resource.MustParse(value)
			} else if key == "cpu" {
				list[corev1.ResourceCPU] = resource.MustParse(value)
			} else {
				list[corev1.ResourceName(key)] = resource.MustParse(value)
			}
		}
	}
	return list, nil
}

// getContainerResources determines if any resources are requested via the spec
// This is for one node, which currently just has one container
func (r *SlurmReconciler) getContainerResources(
	node *api.Node,
) (corev1.ResourceRequirements, error) {

	// memory int, setCPURequest, setCPULimit, setGPULimit int64
	resources := corev1.ResourceRequirements{}

	// Limits
	limits, err := r.getResourceGroup(node.Resources.Limits)
	if err != nil {
		r.Log.Error(err, "🍅️ Resources for Node.Limits")
		return resources, err
	}
	resources.Limits = limits

	// Requests
	requests, err := r.getResourceGroup(node.Resources.Requests)
	if err != nil {
		r.Log.Error(err, "🍅️ Resources for Node.Requests")
		return resources, err
	}
	resources.Requests = requests
	return resources, nil
}

// getPodResources determines if any resources are requested via the spec
func (r *SlurmReconciler) getNodeResources(
	cluster *api.Slurm,
	node api.Node,
) (corev1.ResourceList, error) {

	// memory int, setCPURequest, setCPULimit, setGPULimit int64
	resources, err := r.getResourceGroup(cluster.Spec.Resources)
	if err != nil {
		r.Log.Error(err, "🍅️ Resources for Node.Resources")
		return resources, err
	}
	return resources, nil
}
