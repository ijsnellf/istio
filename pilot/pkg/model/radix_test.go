// Copyright 2018 Istio Authors
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
package model

import (
	"testing"
)

func TestRadix(t *testing.T) {
	r := newRadix()

	contents := []struct {
		config    Config
		hostnames Hostnames
	}{
		{Config{ConfigMeta: ConfigMeta{Name: "cnn"}}, Hostnames{"www.cnn.com", "*.cnn.com", "*.com"}},
		{Config{ConfigMeta: ConfigMeta{Name: "edition_cnn"}}, Hostnames{"edition.cnn.com"}},
		{Config{ConfigMeta: ConfigMeta{Name: "*.co.uk"}}, Hostnames{"*.co.uk"}},
		{Config{ConfigMeta: ConfigMeta{Name: "*"}}, Hostnames{"*"}},
		{Config{ConfigMeta: ConfigMeta{Name: "io"}}, Hostnames{"*.io"}},
		{Config{ConfigMeta: ConfigMeta{Name: "*.preliminary.io"}}, Hostnames{"*.preliminary.io"}},
	}

	for _, content := range contents {
		for _, hostname := range content.hostnames {
			r.Insert(hostname, content.config)
		}
	}

	testCases := []struct {
		in  Hostname
		out Hostnames
	}{
		{"www.cnn.com", Hostnames{"www.cnn.com"}},
		{"money.cnn.com", Hostnames{".cnn.com"}},
		{"edition.cnn.com", Hostnames{"edition.cnn.com"}},
		{"bbc.co.uk", Hostnames{".co.uk"}},
		{"www.wikipedia.org", Hostnames{""}},
		{"*.cnn.com", Hostnames{"www.cnn.com", ".cnn.com", "edition.cnn.com"}},
		{"*.com", Hostnames{".com", ".cnn.com", "www.cnn.com", "edition.cnn.com"}},
		{"*.uk", Hostnames{".co.uk"}},
		{"*.istio.io", Hostnames{".io"}},
		{"*.preliminary.io", Hostnames{".preliminary.io"}},
		{"*.io", Hostnames{".io", ".preliminary.io"}},
		{"nothing.nowhere.net", Hostnames{""}},
		// {"*", Hostnames{"www.cnn.com", ".cnn.com", ".com", "edition.cnn.com", "", ".co.uk"}}, // maintenance burden
	}

	for _, tt := range testCases {
		configs := r.Lookup(tt.in)
		if len(tt.out) != len(configs) {
			t.Errorf("f(%v) -> wanted len()=%v, got len()=%v", tt.in, len(tt.out), len(configs))
			t.Errorf("%#v", configs)
		}
		for _, h := range tt.out {
			if _, ok := configs[h]; !ok {
				t.Errorf("f(%v) -> missing %v", tt.in, h)
			}
		}
	}
}
