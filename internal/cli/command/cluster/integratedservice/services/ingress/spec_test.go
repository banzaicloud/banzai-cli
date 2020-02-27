// Copyright © 2020 Banzai Cloud
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

package ingress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDomain(t *testing.T) {
	testCases := map[string]struct {
		Domain string
		Valid  bool
	}{
		"empty": {
			Domain: "",
			Valid:  false,
		},
		"no dot": {
			Domain: "lorem",
			Valid:  false,
		},
		"whitespace, no dot": {
			Domain: "lorem ipsum",
			Valid:  false,
		},
		"whitespace": {
			Domain: "lorem ipsum.dolor",
			Valid:  false,
		},
		"wildcard": {
			Domain: "*.lorem.ipsum",
			Valid:  true,
		},
		"starts w/ dot": {
			Domain: ".lorem.ipsum",
			Valid:  false,
		},
		"2nd level": {
			Domain: "lorem.ipsum",
			Valid:  true,
		},
		"3rd level": {
			Domain: "lorem.ipsum.dolor",
			Valid:  true,
		},
		"dash in 2nd level": {
			Domain: "lorem-ipsum.dolor",
			Valid:  true,
		},
		"dash in 3rd level": {
			Domain: "lorem.ipsum-dolor",
			Valid:  false,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			err := validateDomain(testCase.Domain)
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
