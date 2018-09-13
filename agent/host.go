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

package agent

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"

	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/topology/graph"
)

// CPUInfo defines host information
type CPUInfo struct {
	CPU        int64
	VendorID   string `json:"VendorID,omitempty"`
	Family     string `json:"Family,omitempty"`
	Model      string `json:"Model,omitempty"`
	Stepping   int64  `json:"Stepping,omitempty"`
	PhysicalID string `json:"PhysicalID,omitempty"`
	CoreID     string `json:"CoreID,omitempty"`
	Cores      int64  `json:"Cores,omitempty"`
	ModelName  string `json:"ModelName,omitempty"`
	Mhz        int64  `json:"Mhz,omitempty"`
	CacheSize  int64  `json:"CacheSize,omitempty"`
	Microcode  string `json:"Microcode,omitempty"`
}

// extractClassAndID to report class and ID from the children node
func extractClassAndID(children map[string]interface{}) (string, string) {
	var foundID string
	var foundClass string
	for objectName, objectValue := range children {
		switch objectValue.(type) {
		case string:
			if objectName == "class" {
				foundClass = objectValue.(string)
			}
			if objectName == "id" {
				foundID = objectValue.(string)
			}
			// Once we discover the class and ID, let's return them
			if (len(foundID) > 0) && (len(foundClass) > 0) {
				return foundClass, foundID
			}
		}
	}
	return "", ""
}

// parseLshw to rename children nodes inside the lshw data structure
func parseLshw(items map[string]interface{}) {
	// For each object of the LSHW structure
	for objectName, objectValue := range items {
		switch objectValue.(type) {
		case []interface{}:
			// Only consider object that have children
			if objectName != "children" {
				continue
			}
			// When we have children
			for _, child := range objectValue.([]interface{}) {
				_, validChildType := child.(map[string]interface{})
				if validChildType {
					// For each child
					class, ID := extractClassAndID(child.(map[string]interface{}))
					// Extract from the child the class and ID to rename it
					if len(class) > 0 {
						if items[class] == nil {
							items[class] = []map[string]interface{}{}
						}
						// Let's create a new child with the name of the class and add an entry with the ID and point it to the actual child
						items[class] = append(items[class].([]map[string]interface{}), map[string]interface{}{ID: child})
						// Delete the actual entry so we renamed it
						delete(items, objectName)
						// Do the same parsing with the current child as he could have some 'children' nodes
						parseLshw(child.(map[string]interface{}))
					}
				}
			}
		}
	}
}

// createRootNode creates a graph.Node based on the host properties and aims to have an unique ID
func createRootNode(g *graph.Graph) (*graph.Node, error) {
	hostID := config.GetString("host_id")
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	m := graph.Metadata{"Name": hostID, "Type": "host", "Hostname": hostname}

	// Fill the metadata from the configuration file
	if configMetadata := config.Get("agent.metadata"); configMetadata != nil {
		configMetadata, ok := common.NormalizeValue(configMetadata).(map[string]interface{})
		if !ok {
			return nil, errors.New("agent.metadata has wrong format")
		}
		for k, v := range configMetadata {
			m[k] = v
		}
	}

	// Retrieves the instance ID from cloud-init
	if buffer, err := ioutil.ReadFile("/var/lib/cloud/data/instance-id"); err == nil {
		m.SetField("InstanceID", strings.TrimSpace(string(buffer)))
	}

	if isolated, err := getIsolatedCPUs(); err == nil {
		m.SetField("IsolatedCPU", isolated)
	}

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	if hostInfo.OS != "" {
		m.SetField("OS", hostInfo.OS)
	}
	if hostInfo.Platform != "" {
		m.SetField("Platform", hostInfo.Platform)
	}
	if hostInfo.PlatformFamily != "" {
		m.SetField("PlatformFamily", hostInfo.PlatformFamily)
	}
	if hostInfo.PlatformVersion != "" {
		m.SetField("PlatformVersion", hostInfo.PlatformVersion)
	}
	if hostInfo.KernelVersion != "" {
		m.SetField("KernelVersion", hostInfo.KernelVersion)
	}
	if hostInfo.VirtualizationSystem != "" {
		m.SetField("VirtualizationSystem", hostInfo.VirtualizationSystem)
	}
	if hostInfo.VirtualizationRole != "" {
		m.SetField("VirtualizationRole", hostInfo.VirtualizationRole)
	}

	var lshwMap map[string]interface{}
	lshw, err := exec.Command("lshw", "-quiet", "-json").Output()
	if err == nil {
		err = json.Unmarshal(lshw, &lshwMap)
		if err == nil {
			parseLshw(lshwMap)
			m.SetField("Hardware", lshwMap)
		}
	} else {
		cpuInfo, err := cpu.Info()
		if err != nil {
			return nil, err
		}
		var cpus []*CPUInfo
		for _, cpu := range cpuInfo {
			c := &CPUInfo{
				CPU:        int64(cpu.CPU),
				VendorID:   cpu.VendorID,
				Family:     cpu.Family,
				Model:      cpu.Model,
				Stepping:   int64(cpu.Stepping),
				PhysicalID: cpu.PhysicalID,
				CoreID:     cpu.CoreID,
				Cores:      int64(cpu.Cores),
				ModelName:  cpu.ModelName,
				Mhz:        int64(cpu.Mhz),
				CacheSize:  int64(cpu.CacheSize),
				Microcode:  cpu.Microcode,
			}
			cpus = append(cpus, c)
		}
		m.SetField("CPU", cpus)
	}

	return g.NewNode(graph.GenID(), m), nil
}
