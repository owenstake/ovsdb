// Copyright 2018 Paul Greenberg (greenpau@outlook.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ovsdb

// OvsDatapath represents an OVS datapath. A datapath is a collection
// of the ports attached to bridges. Each datapath also has associated
// with it a flow table that userspace populates with flows that map
// from keys based on packet headers and metadata to sets of actions.
// Importantly, a datapath is a userspace concept.
type OvsPmd struct {
	NumaId   int
	CoreId   int
	Isolated bool
	usage    float64
}

type OvsRxq struct {
	port    int
	queue   int
	enabled bool
	core    int
	usage   float64
}
