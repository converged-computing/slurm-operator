/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/converged-computing/slurm-operator/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	jobset "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

var (
	backoffLimit = int32(100)

	// We create our own service
	enableDNSHostnames = false
	setHostnameAsFQDN  = true
)

// newJobSet creates the jobset for the slurm
func (r *SlurmReconciler) newJobSet(
	cluster *api.Slurm,
) (*jobset.JobSet, error) {

	// When we have a success policy
	// serverName := cluster.Name + "-server"

	// When suspend is true we have a hard time debugging jobs, so keep false
	suspend := false
	jobs := jobset.JobSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: jobset.JobSetSpec{

			// This would allow pods to be reached by their hostnames!
			// It doesn't work at the moment, but could if we can specify the service name
			// <jobSet.name>-<spec.replicatedJob.name>-<job-index>-<pod-index>.<jobSet.name>-<spec.replicatedJob.name>
			Network: &jobset.Network{
				EnableDNSHostnames: &enableDNSHostnames,
			},

			// The job is successful when the broker job finishes with completed (0)
			//SuccessPolicy: &jobset.SuccessPolicy{
			//	Operator:             jobset.OperatorAny,
			//	TargetReplicatedJobs: []string{serverName},
			//},
			//FailurePolicy: &jobset.FailurePolicy{
			//	MaxRestarts: 0,
			//},

			// This might be the control for child jobs (worker)
			// But I don't think we need this anymore.
			Suspend: &suspend,
		},
	}

	// Get leader login job, the parent in the JobSet
	serverJob, err := r.getJob(cluster, cluster.Spec.Node, 1, "s", true)
	if err != nil {
		r.Log.Error(err, "There was an error getting the server ReplicatedJob")
		return &jobs, err
	}

	// This is the slurm daemon, which looks like a worker
	daemonJob, err := r.getJob(cluster, cluster.Daemon(), 1, "d", true)
	if err != nil {
		r.Log.Error(err, "There was an error getting the daemon ReplicatedJob")
		return &jobs, err
	}

	// Create the database service in the jobset
	dbJob := jobset.ReplicatedJob{}
	if cluster.Spec.DeployDatabase {
		dbJob, err = r.getDatabaseJob(cluster)
		if err != nil {
			r.Log.Error(err, "There was an error getting the database Replicated Job")
			return &jobs, err
		}

	}

	// Create a cluster (JobSet) with workers (required)
	workerNodes := cluster.WorkerNodes()
	workerJob, err := r.getJob(cluster, cluster.WorkerNode(), workerNodes, "w", true)
	if err != nil {
		r.Log.Error(err, "There was an error getting the worker ReplicatedJob")
		return &jobs, err
	}
	if cluster.Spec.DeployDatabase {
		jobs.Spec.ReplicatedJobs = []jobset.ReplicatedJob{serverJob, dbJob, daemonJob, workerJob}
	} else {
		jobs.Spec.ReplicatedJobs = []jobset.ReplicatedJob{serverJob, daemonJob, workerJob}
	}
	ctrl.SetControllerReference(cluster, &jobs, r.Scheme)
	return &jobs, nil
}

// getDatabaseJob gets the database job
func (r *SlurmReconciler) getDatabaseJob(cluster *api.Slurm) (jobset.ReplicatedJob, error) {

	// Default environment
	environment := map[string]string{
		"MYSQL_RANDOM_ROOT_PASSWORD": "yes",
		"MYSQL_DATABASE":             cluster.Spec.Database.DatabaseName,
		"MYSQL_USER":                 cluster.Spec.Database.User,
		"MYSQL_PASSWORD":             cluster.Spec.Database.Password,
	}

	// Update environment with user specifation
	for k, v := range cluster.Spec.Database.Environment {
		environment[k] = v
	}

	podLabels := r.getPodLabels(cluster)
	completionMode := batchv1.IndexedCompletion
	size := int32(1)

	job := jobset.ReplicatedJob{
		Name: "db",
		Template: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: cluster.Namespace,
			},
		},
		// This is the default, but let's be explicit
		Replicas: 1,
	}

	// Create the JobSpec for the job -> Template -> Spec
	jobspec := batchv1.JobSpec{
		BackoffLimit:          &backoffLimit,
		Completions:           &size,
		Parallelism:           &size,
		CompletionMode:        &completionMode,
		ActiveDeadlineSeconds: &cluster.Spec.DeadlineSeconds,

		// Note there is parameter to limit runtime
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    podLabels,
			},
			Spec: corev1.PodSpec{
				// matches the service
				Subdomain:         cluster.ServiceName(),
				SetHostnameAsFQDN: &setHostnameAsFQDN,
				RestartPolicy:     corev1.RestartPolicyOnFailure,
			},
		},
	}

	// Prepare container for database
	// Allow dictating pulling on the level of the node
	pullPolicy := corev1.PullIfNotPresent
	if cluster.Spec.Database.PullAlways {
		pullPolicy = corev1.PullAlways
	}

	// Prepare environment
	envars := []corev1.EnvVar{}
	for key, value := range environment {
		newEnvar := corev1.EnvVar{
			Name:  key,
			Value: value,
		}
		envars = append(envars, newEnvar)
	}

	// Create the containers for the pod (just one for now :)
	containers := []corev1.Container{{
		Name:            "db",
		Image:           cluster.Spec.Database.Image,
		ImagePullPolicy: pullPolicy,
		Stdin:           true,
		TTY:             true,
		Env:             envars,
	}}

	jobspec.Template.Spec.Containers = containers
	job.Template.Spec = jobspec
	return job, nil
}

// getJob creates a job for a main leader (broker) or worker (followers)
func (r *SlurmReconciler) getJob(
	cluster *api.Slurm,
	node api.Node,
	size int32,
	entrypoint string,
	indexed bool,
) (jobset.ReplicatedJob, error) {

	podLabels := r.getPodLabels(cluster)
	completionMode := batchv1.NonIndexedCompletion

	// Is this an indexed job?
	if indexed {
		completionMode = batchv1.IndexedCompletion
	}

	job := jobset.ReplicatedJob{
		Name: entrypoint,
		Template: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      entrypoint,
				Namespace: cluster.Namespace,
			},
		},
		// This is the default, but let's be explicit
		Replicas: 1,
	}

	// Create the JobSpec for the job -> Template -> Spec
	jobspec := batchv1.JobSpec{
		BackoffLimit:          &backoffLimit,
		Completions:           &size,
		Parallelism:           &size,
		CompletionMode:        &completionMode,
		ActiveDeadlineSeconds: &cluster.Spec.DeadlineSeconds,

		// Note there is parameter to limit runtime
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Labels:    podLabels,
			},
			Spec: corev1.PodSpec{
				// matches the service
				SetHostnameAsFQDN: &setHostnameAsFQDN,
				Subdomain:         cluster.ServiceName(),
				Volumes:           getVolumes(cluster),
				RestartPolicy:     corev1.RestartPolicyOnFailure,
			},
		},
	}

	// Do we have a pull secret for the image?
	if node.PullSecret != "" {
		jobspec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: node.PullSecret},
		}
	}

	// Get resources for the node (server or worker)
	resources, err := r.getNodeResources(cluster, node)
	r.Log.Info("👑️ slurm", "Pod.Resources", resources)
	if err != nil {
		r.Log.Error(err, "❌ slurm", "Pod.Resources", resources)
		return job, err
	}
	jobspec.Template.Spec.Overhead = resources

	// Get volume mounts, add on container specific ones
	mounts := getVolumeMounts(cluster)
	containers, err := r.getContainers(
		node,
		mounts,
		entrypoint,
	)
	// Error creating containers
	if err != nil {
		r.Log.Error(err, "❌ slurm", "Pod.Resources", resources)
		return job, err
	}
	jobspec.Template.Spec.Containers = containers
	job.Template.Spec = jobspec
	return job, err
}
