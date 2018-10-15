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

// TODO: handle lookups with wildcards in them?
// TODO: if there are conflicts, pick the oldest config

type hostLookup interface {
	Lookup(hostname Hostname) map[Hostname]Config
}

type reverseRadix struct {
	radix *iradix.Tree
}

func newReverseRadix() *reverseRadix {
	return &reverseRadix{
		radix: iradix.New(),
	}
}

// This function returns the set of the most specific matching virtual services
// for a hostname. It supports wildcards in both the search hostname as well
// as the virtual service hostnames.
//
// To retrieve the most specific matches, we need to define what we consider more or less specific.
// We define the specificity of a match as the amount of the query host that is matched with the 
// config host. A match where more of the query host is matched is considered more specific.
//
// "abc.def" matches both "abc.def" and "*.def"
// "abc.def" has an exact match of all 7 characters of "abc.def", but only 4 characters of "*.def".
// therefore the match of "abc.def" is considered more specific than the match of "*.def"
//
// This definition of specificity becomes important when wildcards are present in both the query
// host and the config host. When the query host contains a wildcard, there can be multiple
// equally specific matches. This is illustrated below with an example:
//
// "*.def" matches "abc.def", "*.def", and "*"
// the match with "abc.def" and "*.def" have equal specificity: both have an exact match of
// 4 characters with the query host.
// the match of "*" is less specific than the other two matches, since the exact match is 0 characters.
// thus, the most specific matches are "abc.def" and "*.def" 
//
// Let's explore some examples, starting with the simplest case and iteratively adding complexity.
//
// Virtual services:
// {"abc.def"} -> A
// {"abc.xyz"} -> B
//
// lookup("abc.def") => A
// lookup("abc.xyz") => B
// lookup("abc.mno") => nil
//
// Now, let's add wildcard virtual service hosts.
//
// Virtual services:
// {"abc.def"} -> A
// {"*.def"} -> B
//
// lookup("abc.def") -> A
// lookup("foo.def") -> B
//
// "abc.def" matches both A and B, but A is a more specific match (more of the lookup hostname is found in A than B).
// For "foo.def", however, the most specific match that exists is B.
//
// Things get extra tricky when we allow wildcards in the lookup host.
//
// Virtual services:
// {"abc.def"} -> A
// {"*.def"} -> B
// {"*"} -> C
//
// lookup("abc.def") -> A (also matches B, C, but those are less specific)
// lookup("foo.def") -> B (also matches C, but that is less specific)
// lookup("*.def") -> A, B (A and B are equally specific. also matches C, but C is less specific than A or B)
// lookup("*.qwerty") -> C (only matches C)
//
//
func (r *reverseRadix) Lookup(hostname Hostname) map[Hostname]Config {
	configs := make(map[Hostname]Config)
	wildcard := strings.Contains(string(hostname), "*")

	//
	if wildcard {
		r.radix.Root().WalkPrefix(r.toKey(hostname), func(k []byte, v interface{}) bool {
			config, _ := v.(Config)
			configs[r.fromKey(k)] = config
			return false
		})
	}

	//
	if !wildcard || len(configs) == 0 {
		k, v, _ := r.radix.Root().LongestPrefix(r.toKey(hostname))
		config, _ := v.(Config)
		configs[r.fromKey(k)] = config
	}

	return configs
}

func (r *reverseRadix) Insert(hostname Hostname, config Config) {
	r.radix, _, _ = r.radix.Insert(r.toKey(hostname), config)
}

// removes wildcard, reverses
func (r *reverseRadix) toKey(hostname Hostname) []byte {
	s := strings.Replace(string(hostname), "*", "", -1)
	data := []byte(s)
	reverse(data)
	return data
}

// unreverses
func (r *reverseRadix) fromKey(key []byte) Hostname {
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
