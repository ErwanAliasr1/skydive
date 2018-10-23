/*
 * Copyright (C) 2018 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package ceph

import (
	"github.com/skydive-project/skydive/topology/graph"
)

// Probe describes graph peering based on MAC address and graph events
type Probe struct {
	graph.DefaultGraphListener
	graph    *graph.Graph
	peers    map[string]*graph.Node
	cluster  CLUSTER
	osds    map[string]string
}

func (p *Probe) onNodeEvent(n *graph.Node) {
	graphCluster(p, n)
	if len(p.cluster.Fsid) == 0 {
		return
	}
	graphOSDs(p, n)
}

// OnNodeUpdated event
func (p *Probe) OnNodeUpdated(n *graph.Node) {
	p.onNodeEvent(n)
}

// OnNodeAdded event
func (p *Probe) OnNodeAdded(n *graph.Node) {
	p.osds[p.cluster.Fsid] = ""
	p.onNodeEvent(n)
}

// OnNodeDeleted event
func (p *Probe) OnNodeDeleted(n *graph.Node) {
	p.osds[p.cluster.Fsid] = ""
}

// Start the probe
func (p *Probe) Start() {
}

// Stop the probe
func (p *Probe) Stop() {
	p.graph.RemoveEventListener(p)
}

// NewAnalyzerProbe update graph to represent a ceph cluster
func NewAnalyzerProbe(g *graph.Graph) *Probe {
	probe := &Probe{
		graph:    g,
		peers:    make(map[string]*graph.Node),
		osds:  make(map[string]string),
	}
	g.AddEventListener(probe)

	return probe
}
