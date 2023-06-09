/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	api "github.com/converged-computing/slurm-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	entrypointSuffix = "-entrypoint"
	configSuffix     = "-config"
)

// GetVolumeMounts returns read only volume for entrypoint scripts, etc.
func getVolumeMounts(cluster *api.Slurm) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      cluster.Name + entrypointSuffix,
			MountPath: "/slurm_operator/",
			ReadOnly:  true,
		},
		// We stage files here, and then will move them over to /etc/slurm
		{
			Name:      cluster.Name + configSuffix,
			MountPath: "/etc/slurm_operator/",
			ReadOnly:  true,
		},
	}
	return mounts
}

// getVolumes for the Indexed Jobs
// !!TODO add slurm.conf and slurmdbd.conf here!!
func getVolumes(cluster *api.Slurm) []corev1.Volume {

	// Runner start scripts
	makeExecutable := int32(0777)
	permission := int32(0600)

	// slurm configuration scripts at /etc/slurm
	slurmConfigs := []corev1.KeyToPath{
		{
			Key:  "slurm.conf",
			Path: "slurm.conf",
			Mode: &permission,
		},
		{
			Key:  "slurmdbd.conf",
			Path: "slurmdbd.conf",
			Mode: &permission,
		},
	}
	// Each of the server and nodes are given the entrypoint scripts
	// Although they won't both use them, this makes debugging easier
	runnerScripts := []corev1.KeyToPath{
		{
			Key:  "start-server",
			Path: "start-s.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  "start-worker",
			Path: "start-w.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  "start-daemon",
			Path: "start-d.sh",
			Mode: &makeExecutable,
		},
	}

	volumes := []corev1.Volume{
		{
			Name: cluster.Name + entrypointSuffix,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{

					// Namespace based on the cluster
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.Name + entrypointSuffix,
					},
					// /slurm_operator/start-worker.sh
					// /slurm_operator/start-server.sh
					Items: runnerScripts,
				},
			},
		},
		{
			Name: cluster.Name + configSuffix,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{

					// Namespace based on the cluster
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.Name + configSuffix,
					},
					// /etc/slurm/slurm.conf
					// /etc/slurm/slurmdbd.conf
					Items: slurmConfigs,
				},
			},
		},
	}
	return volumes
}
