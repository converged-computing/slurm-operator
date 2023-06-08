/*
Copyright 2023 Lawrence Livermore National Security, LLC

	(c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SlurmSpec defines the desired state of slurm
type SlurmSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The generic login node
	Node Node `json:"node"`

	// Slurm dbd "daemon"
	//+optional
	Daemon Node `json:"daemon"`

	// Worker is the worker node spec, does not include login slurmctl or slurmdbd
	// Defaults to be same spec as the server
	//+optional
	Worker Node `json:"worker"`

	// Database is the database service spec
	//+optional
	Database Database `json:"database"`

	// Release of slurm to installed (if sbinary not found in PATH)
	// +kubebuilder:default="19.05.2"
	// +default="19.05.2"
	// +optional
	SlurmVersion string `json:"slurmVersion,omitempty"`

	// Size of the slurm (1 server + (N-1) nodes)
	Size int32 `json:"size"`

	// Interactive mode keeps the cluster running
	// +optional
	Interactive bool `json:"interactive"`

	// Time limit for the job
	// Approximately one year. This cannot be zero or job won't start
	// +kubebuilder:default=31500000
	// +default=31500000
	// +optional
	DeadlineSeconds int64 `json:"deadlineSeconds,omitempty"`

	// Resources include limits and requests
	// +optional
	Resources Resource `json:"resources"`
}

// Database corresponds to the slurm database to use
type Database struct {

	// Image to use for the database
	// We assume we don't need to tweak the command
	// +kubebuilder:default="mariadb:10.10"
	// +default="mariadb:10.10"
	// +optional
	Image string `json:"image"`

	// Default Environment, will be set if not defined here
	// Note that by defalt we set MYSQL_* envars.
	// If you use a different database, be sure to set them all
	// Username and password are set separately below!
	// +optional
	Environment map[string]string `json:"environment"`

	// Database user
	// +kubebuilder:default="slurm"
	// +default="slurm"
	// +optional
	User string `json:"user"`

	// Database password
	// +kubebuilder:default="password"
	// +default="password"
	// +optional
	Password string `json:"password"`

	// Name of the database
	// +kubebuilder:default="slurm_acct_db"
	// +default="slurm_acct_db"
	// +optional
	DatabaseName string `json:"databaseName"`

	// PullAlways will always pull the container
	// +optional
	PullAlways bool `json:"pullAlways"`
}

// Node corresponds to a pod (server or worker)
type Node struct {

	// Image to use for slurm
	// +kubebuilder:default="ghcr.io/converged-computing/slurm"
	// +default="ghcr.io/converged-computing/slurm"
	// +optional
	Image string `json:"image"`

	// Resources include limits and requests
	// +optional
	Resources Resources `json:"resources"`

	// PullSecret for the node, if needed
	// +optional
	PullSecret string `json:"pullSecret"`

	// Command will be honored by a server node
	// +optional
	Command string `json:"command,omitempty"`

	// Commands to run around different parts of the setup
	// +optional
	Commands Commands `json:"commands,omitempty"`

	// Working directory
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// PullAlways will always pull the container
	// +optional
	PullAlways bool `json:"pullAlways"`

	// Ports to be exposed to other containers in the cluster
	// We take a single list of integers and map to the same
	// +optional
	// +listType=atomic
	Ports []int32 `json:"ports"`

	// Key/value pairs for the environment
	// +optional
	Environment map[string]string `json:"environment"`
}

// ContainerResources include limits and requests
type Commands struct {

	// Init runs before anything in both scripts
	// +optional
	Init string `json:"init,omitempty"`
}

// ContainerResources include limits and requests
type Resources struct {

	// +optional
	Limits Resource `json:"limits"`

	// +optional
	Requests Resource `json:"requests"`
}

type Resource map[string]intstr.IntOrString

// Validate the slurm
func (s *Slurm) Validate() bool {
	if s.WorkerNodes() < 1 {
		fmt.Printf("😥️ Slurm cluster must have at least one worker node, Size >= 2.\n")
		return false
	}
	// Ensure we have the default image set
	if s.Spec.Database.Image == "" {
		s.Spec.Database.Image = "mariadb:10.10"
	}

	// Along with a username and password
	if s.Spec.Database.DatabaseName == "" {
		s.Spec.Database.DatabaseName = "slurm_acct_db"
	}
	s.Spec.Database.Password = getRandomToken(s.Spec.Database.Password)
	if s.Spec.Database.User == "" {
		s.Spec.Database.User = "slurm"
	}
	return true
}

// WorkerNodes returns the number of worker nodes
// At this point we've already validated size is >= 1
func (s *Slurm) WorkerNodes() int32 {
	return s.Spec.Size - 1
}

// WorkerNode returns the worker node (if defined) or falls back to the server
func (s *Slurm) WorkerNode() Node {

	// If we don't have a worker spec, copy the parent
	workerNode := s.Spec.Worker
	if reflect.DeepEqual(workerNode, Node{}) {
		workerNode = s.Spec.Node
	}
	return workerNode
}

// getRandomToken returns a requested token, or a generated one
func getRandomToken(requested string) string {
	if requested != "" {
		return requested
	}
	return uuid.New().String()
}

// Daemon falls back to the login node configuatino
func (s *Slurm) Daemon() Node {
	node := s.Spec.Daemon
	if reflect.DeepEqual(node, Node{}) {
		node = s.Spec.Node
	}
	return node
}

// SlurmStatus defines the observed state of slurm
type SlurmStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Slurm is the Schema for the slurms API
type Slurm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlurmSpec   `json:"spec,omitempty"`
	Status SlurmStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// slurmList contains a list of slurm
type SlurmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Slurm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Slurm{}, &SlurmList{})
}
