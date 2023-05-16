// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package capi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/capi/object"
	"github.com/verrazzano/kontainer-engine-driver-ociocne/pkg/variables"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
	"text/template"
)

// toObjects adapts a slice of yaml documents into an object array
func toObjects(yamlDocuments []string) []object.Object {
	var objects []object.Object
	for _, document := range yamlDocuments {
		yamls := strings.Split(document, "---")
		for _, y := range yamls {
			objects = append(objects, object.Object{
				Text: y,
			})
		}
	}

	return objects
}

func loadTextTemplate(o object.Object, variables variables.Variables) ([]unstructured.Unstructured, error) {
	fmt.Printf("+++ loadTextTemplate called . Name = %v +++ \n", variables.Name)
	t, err := template.New("objectText").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).Parse(o.Text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, variables); err != nil {
		fmt.Printf("+++ Error = %v +++ \n", err)
		return nil, err
	}
	templatedBytes := buf.Bytes()
	u, err := toUnstructured(templatedBytes)
	if err != nil {
		return nil, err
	}
	for idx := range u {
		fmt.Printf("+++ Unstructured Object Name = %v +++ \n", object.GVR(&u[idx]))
	}
	return u, nil
}

func toUnstructured(o []byte) ([]unstructured.Unstructured, error) {
	j, err := apiyaml.ToJSON(o)
	if err != nil {
		return nil, err
	}
	obj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, j)
	if err != nil {
		return nil, err
	}
	if u, ok := obj.(*unstructured.Unstructured); ok {
		return []unstructured.Unstructured{*u}, nil
	}
	if us, ok := obj.(*unstructured.UnstructuredList); ok {
		return us.Items, nil
	}

	return nil, errors.New("unknown object type during unstructured serialization")
}
