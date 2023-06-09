/*
Copyright 2022-2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	api "github.com/converged-computing/slurm-operator/api/v1alpha1"
)

// getContainers gets containers for a slurm node
func (r *SlurmReconciler) getContainers(
	node api.Node,
	mounts []corev1.VolumeMount,
	defaultName string,
) ([]corev1.Container, error) {

	// Allow dictating pulling on the level of the node
	pullPolicy := corev1.PullIfNotPresent
	if node.PullAlways {
		pullPolicy = corev1.PullAlways
	}

	// If it's the worker vs. main server determines the entrypoint script
	// The main server takes a custom command to run
	script := fmt.Sprintf("/slurm_operator/start-%s.sh", defaultName)
	command := []string{"/bin/bash", script, node.Command}

	// Create the containers for the pod (just one for now :)
	containers := []corev1.Container{}
	containerName := fmt.Sprintf("%s-node", defaultName)

	// Prepare resources
	resources, err := r.getContainerResources(&node)
	if err != nil {
		r.Log.Error(err, "ERROR getting container resources")
		return containers, err
	}

	// Assemble the container for the node
	newContainer := corev1.Container{
		Name:            containerName,
		Image:           node.Image,
		ImagePullPolicy: pullPolicy,
		WorkingDir:      node.WorkingDir,
		VolumeMounts:    mounts,
		Stdin:           true,
		TTY:             true,
		Resources:       resources,
		Command:         command,
	}

	// Ports and environment
	ports := []corev1.ContainerPort{
		{
			ContainerPort: int32(6717),
			Protocol:      "TCP",
		},
		{
			ContainerPort: int32(6718),
			Protocol:      "TCP",
		},
		{
			ContainerPort: int32(6719),
			Protocol:      "TCP",
		},
	}
	envars := []corev1.EnvVar{}

	// For now we will take ports and have container port == exposed port
	for _, port := range node.Ports {
		newPort := corev1.ContainerPort{
			ContainerPort: int32(port),
			Protocol:      "TCP",
		}
		ports = append(ports, newPort)
	}
	// Add environment variables
	for key, value := range node.Environment {
		newEnvar := corev1.EnvVar{
			Name:  key,
			Value: value,
		}
		envars = append(envars, newEnvar)
	}

	newContainer.Ports = ports
	newContainer.Env = envars
	containers = append(containers, newContainer)
	return containers, nil
}
