package formatting

import (
	"bytes"
	"fmt"
	"github.com/ttacon/chalk"
	"reflect"
	"text/template"
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
	col, err := CustomColumn(name, fieldName, tpl)
	if err != nil {
		panic(err)
	}
	return col
}

func CustomColumn(name, fieldName, tpl string) (*Column, error) {
	template, err := template.New(name).Parse(tpl)
	if err != nil {
		return nil, err
	}
	return &Column{Name: name, Template: template}, nil
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
		panic(fmt.Sprintf("got %T (%s) instead of a slice of struct", slice, s.Kind()))
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

	formattedFields := [][]string{}
	for _, row := range t.Rows {
		formattedRow := []string{}
		for i, column := range t.Columns {
			value := column.FormatField(row)
			formattedRow = append(formattedRow, value)
			len := len(value)
			if len > colWidths[i] {
				colWidths[i] = len
			}
		}
		formattedFields = append(formattedFields, formattedRow)
	}

	//header
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

	//rows
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
