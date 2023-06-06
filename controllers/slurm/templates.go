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
