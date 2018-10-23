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
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/topology/graph"
)

// MON structure extracted from 12.2.5-42.el7cp
type MON struct {
	Name              string `json:"name"`
	Addr              string `json:"addr"`
	Arch              string `json:"arch"`
	CephVersion       string `json:"ceph_version"`
	CPU               string `json:"cpu"`
	Distro            string `json:"distro"`
	DistroDescription string `json:"distro_description"`
	DistroVersion     string `json:"distro_version"`
	Hostname          string `json:"hostname"`
	KernelDescription string `json:"kernel_description"`
	KernelVersion     string `json:"kernel_version"`
	MemSwapKb         string `json:"mem_swap_kb"`
	MemTotalKb        string `json:"mem_total_kb"`
	Os                string `json:"os"`
}

// ReadMons to extract ceph mon metadata
func ReadMons(s *InfoProbe) {
	var mons []MON
	stdout, err := exec.Command("ceph", "mon", "metadata", "-f", "json").Output()
	if err == nil {
		err = json.Unmarshal(stdout, &mons)
		if err == nil {
			var b bytes.Buffer
			e := gob.NewEncoder(&b)
			e.Encode(mons)
			s.g.AddMetadata(s.hostNode, "Software.Ceph.MON.metadata", base64.StdEncoding.EncodeToString(b.Bytes()))
		}
	}
}

// Create a Node per Mon
func graphMon(p *Probe, mon MON) bool {
	lookupNode := p.graph.LookupFirstNode(graph.Metadata{
		"Name": mon.Hostname,
		"Type": "host",
	})

	if lookupNode == nil {
		logging.GetLogger().Errorf("Cannot find any node for host %s", mon.Hostname)
		return false
	}
	monName := fmt.Sprintf("mon.%s", mon.Name)
	nodeName := fmt.Sprintf("%s_%s", mon.Hostname, mon.Name)

	metadata := graph.Metadata{
		"Manager": "ceph",
		"Type":    "MON",
		"Name":    monName,
		"Ceph": map[string]interface{}{
			"MON": mon,
		},
	}

	// Conecting the Mon to the host
	logging.GetLogger().Infof("%s: Adding Mon %s", mon.Hostname, monName)
	containerNode := p.graph.NewNode(graph.Identifier(nodeName), metadata)
	topology.AddOwnershipLink(p.graph, lookupNode, containerNode, nil)
	return true
}

func graphMons(p *Probe, n *graph.Node) bool {
	var mons []MON
	if metadata, _ := n.GetField("Software.Ceph.MON.metadata"); metadata != nil {
		if p.mons[p.cluster.Fsid] == metadata.(string) {
			logging.GetLogger().Infof("Cluster ceph %s is already graphed", p.cluster.Fsid)
			return false
		}
		by, err := base64.StdEncoding.DecodeString(metadata.(string))
		if err != nil {
			logging.GetLogger().Errorf(`failed base64 Decode : %s`, err)
			return false
		}
		b := bytes.Buffer{}
		b.Write(by)
		d := gob.NewDecoder(&b)
		if err := d.Decode(&mons); err != nil {
			logging.GetLogger().Errorf(`failed to Decode : %s`, err)
			return false
		}
		if len(mons) > 0 {
			//logging.GetLogger().Infof("onNodeEvent Received %#v", osds)
			everythingGraphed := true
			for _, mon := range mons {
				if len(mon.Hostname) == 0 {
					continue
				}
				graphed := graphMon(p, mon)
				if (graphed == false) && (everythingGraphed == true) {
					everythingGraphed = false
				}
			}
			if everythingGraphed == false {
				logging.GetLogger().Infof("Mon graphing of cluster %s aborted because of missing nodes", p.cluster.Fsid)
				return false
			}
			// This is the only place where we know the cluster is perfectly rendered
			p.mons[p.cluster.Fsid] = metadata.(string)
			logging.GetLogger().Infof("Ceph cluster %s is rendered", p.cluster.Fsid)
			return true
		}
	}
	return false
}
