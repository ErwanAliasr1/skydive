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
)

// OSD structure extracted from 12.2.5-42.el7cp
type OSD struct {
	ID                         int    `json:"id"`
	Arch                       string `json:"arch"`
	BackAddr                   string `json:"back_addr"`
	BackIface                  string `json:"back_iface"`
	Bluefs                     string `json:"bluefs"`
	BluefsDbAccessMode         string `json:"bluefs_db_access_mode"`
	BluefsDbBlockSize          string `json:"bluefs_db_block_size"`
	BluefsDbDev                string `json:"bluefs_db_dev"`
	BluefsDbDevNode            string `json:"bluefs_db_dev_node"`
	BluefsDbDriver             string `json:"bluefs_db_driver"`
	BluefsDbModel              string `json:"bluefs_db_model"`
	BluefsDbPartitionPath      string `json:"bluefs_db_partition_path"`
	BluefsDbRotational         string `json:"bluefs_db_rotational"`
	BluefsDbSize               string `json:"bluefs_db_size"`
	BluefsDbType               string `json:"bluefs_db_type"`
	BluefsSingleSharedDevice   string `json:"bluefs_single_shared_device"`
	BluestoreBdevAccessMode    string `json:"bluestore_bdev_access_mode"`
	BluestoreBdevBlockSize     string `json:"bluestore_bdev_block_size"`
	BluestoreBdevDev           string `json:"bluestore_bdev_dev"`
	BluestoreBdevDevNode       string `json:"bluestore_bdev_dev_node"`
	BluestoreBdevDriver        string `json:"bluestore_bdev_driver"`
	BluestoreBdevModel         string `json:"bluestore_bdev_model"`
	BluestoreBdevPartitionPath string `json:"bluestore_bdev_partition_path"`
	BluestoreBdevRotational    string `json:"bluestore_bdev_rotational"`
	BluestoreBdevSize          string `json:"bluestore_bdev_size"`
	BluestoreBdevType          string `json:"bluestore_bdev_type"`
	CephVersion                string `json:"ceph_version"`
	CPU                        string `json:"cpu"`
	DefaultDeviceClass         string `json:"default_device_class"`
	Distro                     string `json:"distro"`
	DistroDescription          string `json:"distro_description"`
	DistroVersion              string `json:"distro_version"`
	FrontAddr                  string `json:"front_addr"`
	FrontIface                 string `json:"front_iface"`
	HbBackAddr                 string `json:"hb_back_addr"`
	HbFrontAddr                string `json:"hb_front_addr"`
	Hostname                   string `json:"hostname"`
	JournalRotational          string `json:"journal_rotational"`
	KernelDescription          string `json:"kernel_description"`
	KernelVersion              string `json:"kernel_version"`
	MemSwapKb                  string `json:"mem_swap_kb"`
	MemTotalKb                 string `json:"mem_total_kb"`
	Os                         string `json:"os"`
	OsdData                    string `json:"osd_data"`
	OsdObjectstore             string `json:"osd_objectstore"`
	Rotational                 string `json:"rotational"`
}

// ReadOSD to extract ceph osd metadata
func ReadOSD(s *InfoProbe) {
	var osds []OSD
	stdout, err := exec.Command("ceph", "osd", "metadata", "-f", "json").Output()
	if err == nil {
		err = json.Unmarshal(stdout, &osds)
		if err == nil {
			var b bytes.Buffer
			e := gob.NewEncoder(&b)
			e.Encode(osds)
			s.g.AddMetadata(s.hostNode, "Software.Ceph.OSD.metadata", base64.StdEncoding.EncodeToString(b.Bytes()))
		}
	}
}
