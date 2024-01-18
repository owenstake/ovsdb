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

import (
	"fmt"

	//"github.com/davecgh/go-spew/spew"
	"strconv"
	"strings"
)

// getAppDatapathInterfaces returns a summary of configured datapaths, including statistics
// and a list of connected ports. The port information includes the OpenFlow
// port number, datapath port number, and the type.
//
// Reference: http://www.openvswitch.org/support/dist-docs/ovs-vswitchd.8.txt

func getRxqPmdUsage(db string, sock string, timeout int) ([]*OvsPmd, []*OvsRxq, error) {
	var app Client
	var err error
	cmd := "dpif-netdev/pmd-rxq-show"
	pmds := []*OvsPmd{}
	rxqs := []*OvsRxq{}
	app, err = NewClient(sock, timeout)
	if err != nil {
		app.Close()
		return pmds, rxqs, fmt.Errorf("failed '%s' from %s: %s", cmd, db, err)
	}
	r, err := app.query(cmd, nil)
	if err != nil {
		app.Close()
		return pmds, rxqs, fmt.Errorf("the '%s' command failed for %s: %s", cmd, db, err)
	}
	app.Close()
	response := r.String()
	if response == "" {
		return pmds, rxqs, fmt.Errorf("the '%s' command return no data for %s", cmd, db)
	}
	lines := strings.Split(strings.Trim(response, "\""), "\\n")
	indents := []int{}
	// First, evaluate output depth
	for _, line := range lines {
		indents = append(indents, indentAnalysis(line))
	}
	depth, err := indentDepthAnalysis(indents)
	if err != nil {
		return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output depth analysis", cmd, db)
	}
	// Second, analyze the output
	for _, line := range lines {
		indent := indentAnalysis(line)
		// line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch depth[indent] {
		case 0:
			pmd := &OvsPmd{}
			i := strings.Index(line, ":")
			lineArray := strings.Split(strings.Join(strings.Fields(line[:i]), " "), " ")

			if len(lineArray) != 6 {
				return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. line: 1", cmd, db, line)
			}

			// pmd thread info
			if lineArray[0] != "pmd" || lineArray[1] != "thread" {
				return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. line: 2", cmd, db, line)
			}

			// get numa id
			if lineArray[2] == "numa_id" {
				if value, err := strconv.Atoi(lineArray[3]); err == nil {
					pmd.Numa = value
				}
			} else {
				return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. line: 3", cmd, db, line)
			}

			// get core id
			if lineArray[4] == "core_id" {
				if value, err := strconv.Atoi(lineArray[5]); err == nil {
					pmd.Core = value
				}
			} else {
				return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s", cmd, db, line)
			}
			if pmd != nil {
				pmds = append(pmds, pmd)
			}
		case 1:
			cur_pmd := pmds[len(pmds)-1]
			lineArray := strings.Split(strings.Join(strings.Fields(line), " "), " ")
			rxq := &OvsRxq{}
			switch lineArray[0] {
			case "isolated":
				if lineArray[2] == "true" {
					cur_pmd.Isolated = true
				} else if lineArray[2] == "false" {
					cur_pmd.Isolated = false
				} else {
					return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. flag1", cmd, db, line)
				}
			case "port:":
				rxq.Port = lineArray[1]

				if lineArray[2] == "queue-id:" {
					if value, err := strconv.Atoi(lineArray[3]); err == nil {
						rxq.Queue = value
					} else {
						return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. flag2", cmd, db, line)
					}
				} else {
					return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. flag3", cmd, db, line)
				}

				if lineArray[4] == "(enabled)" {
					rxq.Enabled = true
				} else if lineArray[4] == "(disabled)" {
					rxq.Enabled = false
				} else {
					return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. flag4", cmd, db, line)
				}
				if lineArray[5] == "pmd" && lineArray[6] == "usage:" {
					if value, err := strconv.ParseFloat(lineArray[7], 64); err == nil {
						rxq.Usage = value
					} else {
						return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. flag5", cmd, db, line)
					}
				} else {
					return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage in ouput line %s. array=%v flag6", cmd, db, line, lineArray)
				}
				rxq.Core = cur_pmd.Core
				rxqs = append(rxqs, rxq)
			default:
				return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis: rxq pmd usage", cmd, db)
			}
		default:
			return pmds, rxqs, fmt.Errorf("the '%s' command return for %s failed output analysis. indent %d", cmd, db, indent)
		}
	}
	return pmds, rxqs, nil

}

// GetAppDatapath returns the information about available datapaths.
func (cli *OvsClient) GetRxqPmdUsage(db string) ([]*OvsPmd, []*OvsRxq, error) {
	cli.updateRefs()
	pmds := []*OvsPmd{}
	rxqs := []*OvsRxq{}
	var err error
	switch db {
	case "vswitchd-service":
		pmds, rxqs, err = getRxqPmdUsage(db, cli.Service.Vswitchd.Socket.Control, cli.Timeout)
		if err != nil {
			return pmds, rxqs, err
		}
	default:
		return pmds, rxqs, fmt.Errorf("The '%s' database is unsupported for '%s'", db, "dpif-netdev/pmd-rxq-show")
	}
	return pmds, rxqs, nil
}
