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
	"testing"
)

type row struct {
	Foo string
	Bar string
	Baz int
}

func TestTable(t *testing.T) {
	tests := map[string]struct {
		data     interface{}
		fields   []string
		expected string
	}{
		"struct": {
			data: []row{
				{"foo", "bar", 3},
				{"foofoo", "barbar", 33},
			},
			fields:   []string{"Bar", "Baz", "Foo"},
			expected: "Bar     Baz  Foo   \nbar     3    foo   \nbarbar  33   foofoo",
		},
		"pointer": {
			data: []*row{
				{"foo", "bar", 3},
				{"foofoo", "barbar", 33},
			},
			fields:   []string{"Bar", "Baz", "Foo"},
			expected: "Bar     Baz  Foo   \nbar     3    foo   \nbarbar  33   foofoo",
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			table := NewTable(test.data, test.fields)

			if got := table.Format(false); got != test.expected {
				t.Errorf("unexpected table result\ngot : %s\nwant: %q", got, test.expected)
			}
		})
	}
}
