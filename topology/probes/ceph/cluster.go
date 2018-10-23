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
	"os/exec"

	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology/graph"
)

// CLUSTER structure extracted from 12.2.5-42.el7cp
type CLUSTER struct {
	Fsid   string `json:"fsid"`
	Health struct {
		Checks struct {
			OSDDOWN struct {
				Severity string `json:"severity"`
				Summary  struct {
					Message string `json:"message"`
				} `json:"summary"`
			} `json:"OSD_DOWN"`
			OSDHOSTDOWN struct {
				Severity string `json:"severity"`
				Summary  struct {
					Message string `json:"message"`
				} `json:"summary"`
			} `json:"OSD_HOST_DOWN"`
		} `json:"checks"`
		Status  string `json:"status"`
		Summary []struct {
			Severity string `json:"severity"`
			Summary  string `json:"summary"`
		} `json:"summary"`
		OverallStatus string `json:"overall_status"`
	} `json:"health"`
	ElectionEpoch int      `json:"election_epoch"`
	Quorum        []int    `json:"quorum"`
	QuorumNames   []string `json:"quorum_names"`
	Monmap        struct {
		Epoch    int    `json:"epoch"`
		Fsid     string `json:"fsid"`
		Modified string `json:"modified"`
		Created  string `json:"created"`
		Features struct {
			Persistent []string      `json:"persistent"`
			Optional   []interface{} `json:"optional"`
		} `json:"features"`
		Mons []struct {
			Rank       int    `json:"rank"`
			Name       string `json:"name"`
			Addr       string `json:"addr"`
			PublicAddr string `json:"public_addr"`
		} `json:"mons"`
	} `json:"monmap"`
	Osdmap struct {
		Osdmap struct {
			Epoch          int  `json:"epoch"`
			NumOsds        int  `json:"num_osds"`
			NumUpOsds      int  `json:"num_up_osds"`
			NumInOsds      int  `json:"num_in_osds"`
			Full           bool `json:"full"`
			Nearfull       bool `json:"nearfull"`
			NumRemappedPgs int  `json:"num_remapped_pgs"`
		} `json:"osdmap"`
	} `json:"osdmap"`
	Pgmap struct {
		PgsByState []interface{} `json:"pgs_by_state"`
		NumPgs     int           `json:"num_pgs"`
		NumPools   int           `json:"num_pools"`
		NumObjects int           `json:"num_objects"`
		DataBytes  int           `json:"data_bytes"`
		BytesUsed  int64         `json:"bytes_used"`
		BytesAvail int64         `json:"bytes_avail"`
		BytesTotal int64         `json:"bytes_total"`
	} `json:"pgmap"`
	Fsmap struct {
		Epoch  int           `json:"epoch"`
		ByRank []interface{} `json:"by_rank"`
	} `json:"fsmap"`
	Mgrmap struct {
		Epoch            int           `json:"epoch"`
		ActiveGid        int           `json:"active_gid"`
		ActiveName       string        `json:"active_name"`
		ActiveAddr       string        `json:"active_addr"`
		Available        bool          `json:"available"`
		Standbys         []interface{} `json:"standbys"`
		Modules          []string      `json:"modules"`
		AvailableModules []string      `json:"available_modules"`
		Services         struct {
		} `json:"services"`
	} `json:"mgrmap"`
	Servicemap struct {
		Epoch    int    `json:"epoch"`
		Modified string `json:"modified"`
		Services struct {
		} `json:"services"`
	} `json:"servicemap"`
}

// ReadCluster to extract ceph osd metadata
func ReadCluster(s *InfoProbe) {
	var cluster CLUSTER
	stdout, err := exec.Command("ceph", "-s", "-f", "json").Output()
	if err == nil {
		err = json.Unmarshal(stdout, &cluster)
		if err == nil {
			var b bytes.Buffer
			e := gob.NewEncoder(&b)
			e.Encode(cluster)
			s.g.AddMetadata(s.hostNode, "Software.Ceph.CLUSTER.metadata", base64.StdEncoding.EncodeToString(b.Bytes()))
		}
	}
}

func graphCluster(p *Probe, n *graph.Node) {
	var cluster CLUSTER
	if metadata, _ := n.GetField("Software.Ceph.CLUSTER.metadata"); metadata != nil {
		by, err := base64.StdEncoding.DecodeString(metadata.(string))
		if err != nil {
			logging.GetLogger().Errorf(`failed base64 Decode : %s`, err)
			return
		}
		b := bytes.Buffer{}
		b.Write(by)
		d := gob.NewDecoder(&b)
		if err := d.Decode(&cluster); err != nil {
			logging.GetLogger().Errorf(`failed to Decode : %s`, err)
			return
		}
		p.cluster = cluster
	}
}
