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
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type Outputs map[string]*templates.Output

func NewOutputs() Outputs {
	return make(Outputs)
}

func (o Outputs) AddOutput(name string, output *templates.Output) {
	o[name] = output
}

func (o Outputs) getComponentSource(componentName string) *templates.Source {
	output := o[componentName]
	if output == nil {
		return nil
	}

	return output.Source
}

func (o Outputs) getComponentImage(componentName string) templates.Image {
	output := o[componentName]
	if output == nil {
		return nil
	}
	return output.Image
}

func (o Outputs) getComponentConfig(componentName string) templates.Config {
	output := o[componentName]
	if output == nil {
		return nil
	}
	return output.Config
}

func (o Outputs) GenerateInputs(component *v1alpha1.SupplyChainComponent) *templates.Inputs {
	inputs := &templates.Inputs{
		Sources: map[string]templates.SourceInput{},
		Images:  map[string]templates.ImageInput{},
		Configs: map[string]templates.ConfigInput{},
	}

	for _, referenceSource := range component.Sources {
		source := o.getComponentSource(referenceSource.Component)
		if source != nil {
			inputs.Sources[referenceSource.Name] = templates.SourceInput{
				URL:      source.URL,
				Revision: source.Revision,
				Name:     referenceSource.Name,
			}
		}
	}

	for _, referenceImage := range component.Images {
		image := o.getComponentImage(referenceImage.Component)
		if image != nil {
			inputs.Images[referenceImage.Name] = templates.ImageInput{
				Image: image,
				Name:  referenceImage.Name,
			}
		}
	}

	for _, referenceConfig := range component.Configs {
		config := o.getComponentConfig(referenceConfig.Component)
		if config != nil {
			inputs.Configs[referenceConfig.Name] = templates.ConfigInput{
				Config: config,
				Name:   referenceConfig.Name,
			}
		}
	}

	return inputs
}
