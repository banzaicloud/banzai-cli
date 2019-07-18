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

package formatting

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/ttacon/chalk"
)

type Column struct {
	Name      string
	Template  *template.Template
	MaxLength int
}

type Table struct {
	Columns   []Column
	Rows      []interface{}
	Separator string
}

const ellipsis = "…"

func (c *Column) FormatFieldOrError(data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	tpl := c.Template
	err := tpl.Execute(buf, data)

	return buf.String(), err
}

func (c *Column) FormatField(data interface{}) string {
	result, err := c.FormatFieldOrError(data)
	if err != nil {
		return fmt.Sprintf("#(%v)", err)
	}

	return trunc(result, c.MaxLength)
}

func trunc(s string, length int) string {
	if length > 0 && len(s) > length {
		return s[0:length-len(ellipsis)] + ellipsis
	}

	return s
}

func NewColumn(name string) *Column {
	return NamedColumn(name, name)
}

func NamedColumn(name, fieldName string) *Column {
	tpl := fmt.Sprintf("{{.%s}}", fieldName)
	col, err := CustomColumn(name, tpl)
	if err != nil {
		panic(err)
	}

	return col
}

func CustomColumn(name, tpl string) (*Column, error) {
	parsedTemplate, err := template.New(name).Parse(tpl)
	if err != nil {
		return nil, err
	}

	return &Column{Name: name, Template: parsedTemplate}, nil
}

func NewTable(data interface{}, fields []string) *Table {
	columns := make([]Column, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, *NewColumn(field))
	}

	slice := asSlice(data)

	return &Table{Columns: columns, Rows: slice, Separator: "  "}
}

func asSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		slice = []interface{}{s}
		s = reflect.ValueOf(slice)
	}

	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func (t *Table) Format(color bool) string {
	colWidths := make([]int, len(t.Columns))
	for i, column := range t.Columns {
		colWidths[i] = len(column.Name)
	}

	formattedFields := make([][]string, len(t.Rows))
	for i, row := range t.Rows {
		formattedRow := make([]string, len(t.Columns))

		for i, column := range t.Columns {
			value := column.FormatField(row)
			formattedRow[i] = value

			if len := len(value); len > colWidths[i] {
				colWidths[i] = len
			}
		}

		formattedFields[i] = formattedRow
	}

	// header
	out := ""
	for i, column := range t.Columns {
		if i > 0 {
			out += t.Separator
		}

		out += fmt.Sprintf("%- *s", colWidths[i], column.Name)
	}
	if color {
		out = chalk.Bold.TextStyle(out)
	}

	// rows
	for _, row := range formattedFields {
		out += "\n"

		for i, field := range row {
			if i > 0 {
				out += t.Separator
			}

			out += fmt.Sprintf("%- *s", colWidths[i], field)
		}
	}

	return out
}
