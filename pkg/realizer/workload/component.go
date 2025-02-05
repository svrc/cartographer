// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workload

import (
	"context"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ComponentRealizer
type ComponentRealizer interface {
	Do(ctx context.Context, component *v1alpha1.SupplyChainComponent, supplyChainName string, outputs Outputs) (*templates.Output, error)
}

type componentRealizer struct {
	workload *v1alpha1.Workload
	repo     repository.Repository
}

func NewComponentRealizer(workload *v1alpha1.Workload, repo repository.Repository) ComponentRealizer {
	return &componentRealizer{
		workload: workload,
		repo:     repo,
	}
}

func (r *componentRealizer) Do(ctx context.Context, component *v1alpha1.SupplyChainComponent, supplyChainName string, outputs Outputs) (*templates.Output, error) {
	template, err := r.repo.GetClusterTemplate(component.TemplateRef)
	if err != nil {
		return nil, GetClusterTemplateError{
			Err:         err,
			TemplateRef: component.TemplateRef,
		}
	}

	labels := map[string]string{
		"carto.run/workload-name":             r.workload.Name,
		"carto.run/workload-namespace":        r.workload.Namespace,
		"carto.run/cluster-supply-chain-name": supplyChainName,
		"carto.run/component-name":            component.Name,
		"carto.run/template-kind":             template.GetKind(),
		"carto.run/cluster-template-name":     template.GetName(),
	}

	inputs := outputs.GenerateInputs(component)
	workloadTemplatingContext := map[string]interface{}{
		"workload": r.workload,
		"params":   templates.ParamsBuilder(template.GetDefaultParams(), component.Params),
		"sources":  inputs.Sources,
		"images":   inputs.Images,
		"configs":  inputs.Configs,
	}
	if inputs.OnlyConfig() != nil {
		workloadTemplatingContext["config"] = inputs.OnlyConfig()
	}
	if inputs.OnlyImage() != nil {
		workloadTemplatingContext["image"] = inputs.OnlyImage()
	}
	if inputs.OnlySource() != nil {
		workloadTemplatingContext["source"] = inputs.OnlySource()
	}

	stampContext := templates.StamperBuilder(r.workload, workloadTemplatingContext, labels)
	stampedObject, err := stampContext.Stamp(ctx, template.GetResourceTemplate())
	if err != nil {
		return nil, StampError{
			Err:       err,
			Component: component,
		}
	}

	err = r.repo.EnsureObjectExistsOnCluster(stampedObject, true)
	if err != nil {
		return nil, ApplyStampedObjectError{
			Err:           err,
			StampedObject: stampedObject,
		}
	}

	output, err := template.GetOutput(stampedObject)
	if err != nil {
		return nil, RetrieveOutputError{
			Err:       err,
			component: component,
		}
	}

	return output, nil
}
