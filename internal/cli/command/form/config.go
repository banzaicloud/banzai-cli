// Copyright © 2018 Banzai Cloud
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

package form

import (
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/ghodss/yaml"
)

// ConfigFile Parsed form config file
type ConfigFile struct {
	Form      []*Group          `json:"form"`
	Templates map[string]string `json:"templates"`
}

func (f ConfigFile) validate() error {
	if f.Form == nil {
		return fmt.Errorf("validate %v: form definition is required", f)
	}

	for _, g := range f.Form {
		err := g.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Group Form group
type Group struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Link        *string  `json:"link,omitempty"`
	Fields      []*Field `json:"fields,omitempty"`
}

func (g Group) validate() error {
	if g.Name == "" {
		return fmt.Errorf("validate %v: group name is required", g)
	}

	if len(g.Fields) == 0 {
		return fmt.Errorf("validate %v: fields cannot be empty", g)
	}

	for _, f := range g.Fields {
		err := f.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Field Form field
type Field struct {
	Key              string                  `json:"key"`
	Label            string                  `json:"label"`
	ControlType      string                  `json:"controlType,omitempty"`
	ControlGroupType string                  `json:"controlGroupType,omitempty"`
	Options          []interface{}           `json:"options,omitempty"`
	Required         *bool                   `json:"required,omitempty"`
	Hidden           *bool                   `json:"hidden,omitempty"`
	Disabled         *bool                   `json:"disabled,omitempty"`
	Placeholder      *string                 `json:"placeholder,omitempty"`
	Description      *string                 `json:"description,omitempty"`
	MinLength        *int                    `json:"minLength,omitempty"`
	MaxLength        *int                    `json:"maxLength,omitempty"`
	Min              *float64                `json:"min,omitempty"`
	Max              *float64                `json:"max,omitempty"`
	Pattern          *string                 `json:"pattern,omitempty"`
	DefaultValue     interface{}             `json:"default,omitempty"`
	Value            interface{}             `json:"value,omitempty"`
	ShowIf           *map[string]interface{} `json:"showIf,omitempty"`
}

func (f Field) validate() error {
	if f.Key == "" {
		return fmt.Errorf("validate %v: field key is required", f.Key)
	}

	if f.ControlType != "" && f.Label == "" {
		return fmt.Errorf("validate %v: field label is required", f.Key)
	}

	if f.ControlType == "" && f.ControlGroupType == "" {
		return fmt.Errorf("validate %s: either controlType or controlGroupType is required", f.Key)
	}

	return nil
}

func readConfig(filename string) (file ConfigFile, err error) {
	filename, raw, err := utils.ReadFileOrStdin(filename)
	if err != nil {
		return file, errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	err = yaml.Unmarshal(raw, &file)
	if err != nil {
		return file, errors.WrapIf(err, "invalid format")
	}

	err = file.validate()
	if err != nil {
		return file, err
	}

	return file, nil
}

func writeConfig(fileName string, file ConfigFile) error {
	config, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	var fi *os.File
	if fileName == "-" {
		fi = os.Stdout
	} else {
		fi, err = os.Create(fileName)
		if err != nil {
			return err
		}

		defer fi.Close()
	}

	_, err = fi.Write(config)
	if err != nil {
		return err
	}

	return nil
}
