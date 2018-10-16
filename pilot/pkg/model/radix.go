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
	"github.com/hashicorp/go-immutable-radix"
	"strings"
)

// TODO: if there are conflicts, pick the oldest config

type hostLookup interface {
	Lookup(hostname Hostname) map[Hostname]Config
}

type radix struct {
	radix *iradix.Tree
}

func newRadix() *radix {
	return &radix{
		radix: iradix.New(),
	}
}

// This function returns the set of the most specific matching configs
// for a hostname. It supports wildcards in both the query hostname as well
// as the config hostnames.
//
// To retrieve the most specific matches, we need to define what we consider more or less specific.
// We define the specificity of a match as the amount of the query host that is matched with the
// config host. A match where more of the query host is matched is considered more specific.
//
// Consider the query hostname "abc.def" and config hostnames "abc.def" and "*.def":
// - The query hostname "abc.def" matches both "abc.def" and "*.def"
// - The query hostname "abc.def" has an exact match of all 7 characters of "abc.def", but only 4
//   characters of "*.def". Therefore the match of "abc.def" is considered more specific than
//   the match of "*.def".
//
// This definition of specificity becomes important when wildcards are present in both the query
// host and the config host. When the query host contains a wildcard, there can be multiple
// equally specific matches. This is illustrated below with an example:
//
// The query host "*.def" matches "abc.def", "*.def", and "*"
// - the match with "abc.def" and "*.def" have equal specificity: both have an exact match of
//   4 characters with the query host.
// - the match of "*" is less specific than the other two matches, since the exact match is 0 characters.
//   thus, the most specific matches are "abc.def" and "*.def"
//
// This function uses a radix to implement the behavior described above.
func (r *radix) Lookup(hostname Hostname) map[Hostname]Config {
	configs := make(map[Hostname]Config)
	wildcard := strings.Contains(string(hostname), "*")

	// If a wildcard is present in the query hostname there may be multiple equally specific matches,
	// so we attempt to walk every config hostname under this prefix.
	if wildcard {
		r.radix.Root().WalkPrefix(r.toKey(hostname), func(k []byte, v interface{}) bool {
			config, _ := v.(Config)
			configs[r.fromKey(k)] = config
			return false
		})
	}

	// If the query hostname has no wildcard, or there were no configs under the prefix, we get the
	// longest matching prefix for this query hostname.
	if !wildcard || len(configs) == 0 {
		k, v, _ := r.radix.Root().LongestPrefix(r.toKey(hostname))
		config, _ := v.(Config)
		configs[r.fromKey(k)] = config
	}

	return configs
}

func (r *radix) Insert(hostname Hostname, config Config) {
	r.radix, _, _ = r.radix.Insert(r.toKey(hostname), config)
}

// Strips the wildcard character '*' and stores the hostname in the radix in reversed character order.
func (r *radix) toKey(hostname Hostname) []byte {
	s := strings.Replace(string(hostname), "*", "", -1)
	data := []byte(s)
	reverse(data)
	return data
}

// Unreverses the hostname.
func (r *radix) fromKey(key []byte) Hostname {
	data := make([]byte, len(key))
	copy(data, key)
	reverse(data)
	return Hostname(data)
}

func reverse(data []byte) {
	for i := 0; i < len(data)/2; i++ {
		data[i], data[len(data)-i-1] = data[len(data)-i-1], data[i]
	}
}
