/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"bytes"
	"fmt"
	"text/template"

	api "github.com/converged-computing/slurm-operator/api/v1alpha1"

	_ "embed"
)

//go:embed templates/server.sh
var startServerTemplate string

//go:embed templates/worker.sh
var startWorkerTemplate string

//go:embed templates/daemon.sh
var startDaemonTemplate string

//go:embed templates/components.sh
var startComponents string

//go:embed templates/slurm.conf
var slurmConf string

//go:embed templates/slurmdbd.conf
var slurmDbdConf string

// NodeTemplate populates a node entrypoint
type NodeTemplate struct {
	Node api.Node
	Spec api.SlurmSpec
}

type ConfigTemplate struct {
	Spec         api.SlurmSpec
	DaemonHost   string
	ControlHost  string
	DatabaseHost string
	Hostlist     string
}

// combineTemplates into one "start"
func combineTemplates(listing ...string) (t *template.Template, err error) {
	t = template.New("start")

	for i, templ := range listing {
		_, err = t.New(fmt.Sprint("_", i)).Parse(templ)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

// generateWorkerScript generates the main script to start everything up!
func generateScript(cluster *api.Slurm, node api.Node, startTemplate string) (string, error) {

	nt := NodeTemplate{
		Node: node,
		Spec: cluster.Spec,
	}

	// Wrap the named template to identify it later
	startTemplate = `{{define "start"}}` + startTemplate + "{{end}}"

	// We assemble different strings (including the components) into one!
	t, err := combineTemplates(startComponents, startTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := t.ExecuteTemplate(&output, "start", nt); err != nil {
		return "", err
	}
	return output.String(), nil
}

// generateHostlist for a specific size given the cluster namespace and a size
// slurm-sample-w-0-0.slurm-service.slurm-operator.svc.cluster.local
func generateHostlist(cluster *api.Slurm) string {
	hosts := ""

	for i := 0; i < int(cluster.WorkerNodes()); i++ {
		if hosts == "" {
			hosts = fmt.Sprintf("%s-w-0-%d.%s.%s.svc.cluster.local", cluster.Name, i, serviceName, cluster.Namespace)
		} else {
			hosts = fmt.Sprintf("%s,%s-w-0-%d.%s.%s.svc.cluster.local", hosts, cluster.Name, i, serviceName, cluster.Namespace)
		}
	}
	return hosts
}

func generateConfig(cluster *api.Slurm, startTemplate string) (string, error) {

	// control: slurm-sample-<x>-0-0.slurm-service.slurm-operator.svc.cluster.local
	control := fmt.Sprintf("%s-s-0-0.%s.%s.svc.cluster.local", cluster.Name, serviceName, cluster.Namespace)
	database := fmt.Sprintf("%s-db-0-0.%s.%s.svc.cluster.local", cluster.Name, serviceName, cluster.Namespace)
	daemon := fmt.Sprintf("%s-d-0-0.%s.%s.svc.cluster.local", cluster.Name, serviceName, cluster.Namespace)

	ct := ConfigTemplate{
		Spec:         cluster.Spec,
		ControlHost:  control,
		DatabaseHost: database,
		DaemonHost:   daemon,
		Hostlist:     generateHostlist(cluster),
	}

	// Wrap the named template to identify it later
	startTemplate = `{{define "start"}}` + startTemplate + "{{end}}"

	// We assemble different strings (including the components) into one!
	t, err := combineTemplates(startComponents, startTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := t.ExecuteTemplate(&output, "start", ct); err != nil {
		return "", err
	}
	return output.String(), nil
}
