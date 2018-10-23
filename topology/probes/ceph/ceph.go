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
	"time"

	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology/graph"
)

//GetCephMetadata to read all the possible ceph information for the host
func GetCephMetadata(s *InfoProbe) {
	ReadCluster(s)
	ReadOSD(s)
	ReadMons(s)
}

// InfoProbe describes a ceph cluster
type InfoProbe struct {
	graph.DefaultGraphListener
	g        *graph.Graph
	hostNode *graph.Node // graph node of the running host
}

// Start the flow Probe
func (s *InfoProbe) Start() {
	logging.GetLogger().Infof("Starting Ceph capture")
	GetCephMetadata(s)
	go func() {
		seconds := config.GetInt("agent.topology.socketinfo.host_update")
		ticker := time.NewTicker(time.Duration(seconds) * 15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				GetCephMetadata(s)
			}
		}
	}()
}

// Stop the flow Probe
func (s *InfoProbe) Stop() {
	logging.GetLogger().Infof("Stopping Ceph capture")
}

// NewAgentProbe create a new Ceph Probe
func NewAgentProbe(g *graph.Graph, hostNode *graph.Node) (*InfoProbe, error) {
	return &InfoProbe{
		g:        g,
		hostNode: hostNode,
	}, nil
}
