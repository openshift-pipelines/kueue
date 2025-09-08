/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generic

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	configapi "sigs.k8s.io/kueue/apis/config/v1beta1"
)

// ConfigManager manages external framework configurations for generic adapters
type ConfigManager struct {
	configs map[string]configapi.MultiKueueExternalFramework
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]configapi.MultiKueueExternalFramework),
	}
}

// LoadConfigurations loads and validates external framework configurations
func (cm *ConfigManager) LoadConfigurations(configs []configapi.MultiKueueExternalFramework) error {
	cm.configs = make(map[string]configapi.MultiKueueExternalFramework)
	var errors []error

	for _, config := range configs {
		if err := cm.validateConfig(config); err != nil {
			klog.Errorf("Invalid external framework configuration: %v", err)
			errors = append(errors, fmt.Errorf("config %s: %w", config.Name, err))
			continue // Skip invalid configurations but continue loading others
		}

		// Parse the GVK from the name field using schema.ParseKindArg
		gvk, _ := schema.ParseKindArg(config.Name)
		if gvk == nil {
			err := fmt.Errorf("invalid GVK format in configuration: %s", config.Name)
			klog.Error(err)
			errors = append(errors, err)
			continue
		}

		// Store the configuration
		cm.configs[gvk.String()] = config
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d configuration errors (see logs for details)", len(errors))
	}
	return nil
}

// GetAdapter returns a generic adapter for the given GVK if configured
func (cm *ConfigManager) GetAdapter(gvk schema.GroupVersionKind) *genericAdapter {
	_, exists := cm.configs[gvk.String()]
	if !exists {
		return nil
	}

	return &genericAdapter{
		gvk: gvk,
	}
}

// GetAllAdapters returns all configured generic adapters
func (cm *ConfigManager) GetAllAdapters() []*genericAdapter {
	adapters := make([]*genericAdapter, 0, len(cm.configs))
	for _, config := range cm.configs {
		// Parse the GVK string back to schema.GroupVersionKind
		gvk, _ := schema.ParseKindArg(config.Name)
		if gvk == nil {
			klog.Errorf("Failed to parse GVK string %s", config.Name)
			continue
		}
		adapters = append(adapters, &genericAdapter{
			gvk: *gvk,
		})
	}
	return adapters
}

// validateConfig validates an external framework configuration
func (cm *ConfigManager) validateConfig(config configapi.MultiKueueExternalFramework) error {
	if config.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate the GVK format using schema.ParseKindArg
	gvk, _ := schema.ParseKindArg(config.Name)
	if gvk == nil {
		return fmt.Errorf("invalid GVK format '%s'", config.Name)
	}

	return nil
}
