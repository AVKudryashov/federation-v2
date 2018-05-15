/*
Copyright 2018 The Federation v2 Authors.

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

package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kubernetes-sigs/federation-v2/pkg/controller/util"
	"github.com/kubernetes-sigs/federation-v2/pkg/federatedtypes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func NewTestObjects(typeConfig federatedtypes.FederatedTypeConfig, namespace string, clusterNames []string) (template, placement, override *unstructured.Unstructured, err error) {
	path := fixturePath()

	filenameTemplate := filepath.Join(path, fmt.Sprintf("%s-%%s.yaml", strings.ToLower(typeConfig.Target.Kind)))

	templateFilename := fmt.Sprintf(filenameTemplate, "template")
	template, err = fileToObj(templateFilename)
	if err != nil {
		return nil, nil, nil, err
	}
	template.SetNamespace(namespace)
	template.SetName("")
	template.SetGenerateName("test-crud-")

	placementFilename := filepath.Join(path, "placement.yaml")
	placement, err = fileToObj(placementFilename)
	if err != nil {
		return nil, nil, nil, err
	}
	placementConfig := typeConfig.Placement
	placement.SetNamespace(namespace)
	placement.SetKind(placementConfig.Kind)
	placement.SetAPIVersion(fmt.Sprintf("%s/%s", placementConfig.Group, placementConfig.Version))
	util.SetClusterNames(placement, clusterNames)

	if typeConfig.Override != nil {
		overrideFilename := fmt.Sprintf(filenameTemplate, "override")
		override, err = fileToObj(overrideFilename)
		if err != nil {
			return nil, nil, nil, err
		}
		override.SetNamespace(namespace)
		overrideSlice, ok := unstructured.NestedSlice(override.Object, "spec", "overrides")
		if !ok {
			return nil, nil, nil, fmt.Errorf("Unable to set override for %q", typeConfig.Template.Kind)
		}
		targetOverride := overrideSlice[0].(map[string]interface{})
		targetOverride["clusterName"] = clusterNames[0]
		overrideSlice[0] = targetOverride
		unstructured.SetNestedSlice(override.Object, overrideSlice, "spec", "overrides")
	}

	return template, placement, override, nil
}

func fixturePath() string {
	// Get the directory of the current executable
	_, filename, _, _ := runtime.Caller(0)
	commonPath := filepath.Dir(filename)
	return filepath.Join(commonPath, "fixtures")
}

func fileToObj(filename string) (*unstructured.Unstructured, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readerToObj(f)
}

func readerToObj(r io.Reader) (*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLToJSONDecoder(r)
	obj := &unstructured.Unstructured{}
	err := decoder.Decode(obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
