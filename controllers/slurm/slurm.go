/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	jobset "sigs.k8s.io/jobset/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/converged-computing/slurm-operator/api/v1alpha1"
)

// A slurm is one or more workers plus a main server

// newslurm creates a new slurm
func (r *SlurmReconciler) ensureslurm(
	ctx context.Context,
	cluster *api.Slurm,
) (ctrl.Result, error) {

	// Add entrypoint config maps and those in /etc/slurm
	_, result, err := r.ensureConfigMap(ctx, cluster, "entrypoint", cluster.Name+entrypointSuffix)
	if err != nil {
		return result, err
	}
	_, result, err = r.ensureConfigMap(ctx, cluster, "slurmconf", cluster.Name+configSuffix)
	if err != nil {
		return result, err
	}

	// Create headless service for the slurm cluster
	selector := map[string]string{"cluster-name": cluster.Name}
	result, err = r.exposeServices(ctx, cluster, serviceName, selector)
	if err != nil {
		return result, err
	}

	// Create the batch job that brings it all together!
	// A batchv1.Job can hold a spec for containers that use the configs we just made
	_, result, err = r.getCluster(ctx, cluster)
	if err != nil {
		return result, err
	}
	// And we re-queue so the Ready condition triggers next steps!
	return ctrl.Result{Requeue: true}, nil
}

// getExistingJob gets an existing job that matches our CRD
func (r *SlurmReconciler) getExistingJob(
	ctx context.Context,
	cluster *api.Slurm,
) (*jobset.JobSet, error) {

	existing := &jobset.JobSet{}
	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		existing,
	)
	return existing, err
}

// getCluster does an actual check if we have a jobset in the namespace
func (r *SlurmReconciler) getCluster(
	ctx context.Context,
	cluster *api.Slurm,
) (*jobset.JobSet, ctrl.Result, error) {

	// Look for an existing job
	existing, err := r.getExistingJob(ctx, cluster)

	// Create a new job if it does not exist
	if err != nil {

		if errors.IsNotFound(err) {
			job, err := r.newJobSet(cluster)
			if err != nil {
				r.Log.Error(
					err,
					"Failed to create new slurm JobSet",
					"Namespace:", job.Namespace,
					"Name:", job.Name,
				)
				// If there is an error, return the existing (empty)
				return existing, ctrl.Result{}, err
			}

			r.Log.Info(
				"✨ Creating a new slurm JobSet ✨",
				"Namespace:", job.Namespace,
				"Name:", job.Name,
			)

			err = r.Client.Create(ctx, job)
			if err != nil {
				r.Log.Error(
					err,
					"Failed to create new slurm JobSet",
					"Namespace:", job.Namespace,
					"Name:", job.Name,
				)
				return existing, ctrl.Result{}, err
			}
			return job, ctrl.Result{}, err

		} else if err != nil {
			r.Log.Error(err, "Failed to get slurm JobSet")
			return existing, ctrl.Result{}, err
		}

	} else {
		r.Log.Info(
			"🎉 Found existing slurm JobSet 🎉",
			"Namespace:", existing.Namespace,
			"Name:", existing.Name,
		)
	}
	return existing, ctrl.Result{}, err
}

// getConfigMap generates the config map, when does not exist
func (r *SlurmReconciler) getConfigMap(
	ctx context.Context,
	cluster *api.Slurm,
	configName string,
	configFullName string,
) (*corev1.ConfigMap, ctrl.Result, error) {

	// Data for the config map
	data := map[string]string{}
	cm := &corev1.ConfigMap{}

	// This is currently the only config we support
	if configName == "entrypoint" {

		// Generate data for both the start-server.sh and start-worker.sh
		serverStart, err := generateScript(cluster, cluster.Spec.Login, startServerTemplate)
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		workerStart, err := generateScript(cluster, cluster.WorkerNode(), startWorkerTemplate)
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		daemonStart, err := generateScript(cluster, cluster.WorkerNode(), startDaemonTemplate)
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		data["start-server"] = serverStart
		data["start-worker"] = workerStart
		data["start-daemon"] = daemonStart
	}

	if configName == "slurmconf" {

		slurmconfig, err := generateConfig(cluster, slurmConf)
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		dbdconfig, err := generateConfig(cluster, slurmDbdConf)
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		data["slurm.conf"] = slurmconfig
		data["slurmdbd.conf"] = dbdconfig
	}

	// Create the config map with respective data!
	cm = r.createConfigMap(cluster, configFullName, data)

	// Actually create it
	err := r.Create(ctx, cm)
	if err != nil {
		r.Log.Error(
			err, "❌ Failed to create slurm ConfigMap",
			"Type", configName,
			"Namespace", cm.Namespace,
			"Name", (*cm).Name,
		)
		return cm, ctrl.Result{}, err
	}

	// Successful - return and requeue
	return cm, ctrl.Result{Requeue: true}, nil
}

// createConfigMap generates a config map with some kind of data
func (r *SlurmReconciler) createConfigMap(
	cluster *api.Slurm,
	configName string,
	data map[string]string,
) *corev1.ConfigMap {

	// Create the config map with respective data!
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: cluster.Namespace,
		},
		Data: data,
	}
	// Finally create the config map
	r.Log.Info(
		"✨ Creating slurm ConfigMap ✨",
		"Type", configName,
		"Namespace", cm.Namespace,
		"Name", cm.Name,
	)
	// Show in the logs
	fmt.Println(cm.Data)
	ctrl.SetControllerReference(cluster, cm, r.Scheme)
	return cm
}

// ensureConfigMap ensures we've generated the read only entrypoints
func (r *SlurmReconciler) ensureConfigMap(
	ctx context.Context,
	cluster *api.Slurm,
	configName string,
	configFullName string,
) (*corev1.ConfigMap, ctrl.Result, error) {

	// Look for the config map by name
	existing := &corev1.ConfigMap{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      configFullName,
			Namespace: cluster.Namespace,
		},
		existing,
	)

	if err != nil {

		// Case 1: not found yet, and hostfile is ready (recreate)
		if errors.IsNotFound(err) {
			return r.getConfigMap(ctx, cluster, configName, configFullName)

		} else if err != nil {
			r.Log.Error(err, "Failed to get slurm ConfigMap")
			return existing, ctrl.Result{}, err
		}

	} else {
		r.Log.Info(
			"🎉 Found existing slurm ConfigMap",
			"Type", configName,
			"Namespace", existing.Namespace,
			"Name", existing.Name,
		)
	}
	return existing, ctrl.Result{}, err
}
