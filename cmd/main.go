/// Copyright 2020 DataStax, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	//	"flag"
	"fmt"
	"os"
)

//help outputs the standard help
func help() string {
	//TODO dynamically generate from flags
	return `usage: sperf [-h] [-p] [-v] [-d DIAG_DIR] [-s SYSTEM_LOG_PREFIX]
             [-l DEBUG_LOG_PREFIX] [-o OUTPUT_LOG_PREFIX]
             [-n NODE_INFO_PREFIX] [-c CFSTATS_PREFIX] [-b BLOCK_DEV_PREFIX]
             {core,search,sysbottle,ttop} ...

Sperf provides a number of useful reports from the diagtarball, iostat and the nodes themselves.

optional arguments:
  -h, --help            show this help message and exit
  -p, --progress        shows file progress to show how long it takes to process each file
  -v, --debug           shows debug output. Useful for bug reports and diagnosing issues with sperf
  -d DIAG_DIR, --diagdir DIAG_DIR
                        where the diag tarball directory is exported, should be where the nodes folder is located (default ".")
  -s SYSTEM_LOG_PREFIX, --system_log_prefix SYSTEM_LOG_PREFIX
                        if system.log in the diag tarball has an oddball name, can still look based on this prefix (default "system.log")
  -l DEBUG_LOG_PREFIX, --debug_log_prefix DEBUG_LOG_PREFIX
                        if debug.log in the diag tarball has an oddball name, can still look based on this prefix (default "debug.log")
  -o OUTPUT_LOG_PREFIX, --output_log_prefix OUTPUT_LOG_PREFIX
                        if output.log in the diag tarball has an oddball name, can still look based on this prefix (default "output.log")
  -n NODE_INFO_PREFIX, --node_info_prefix NODE_INFO_PREFIX
                        if node_info.json in the diag tarball has an oddball name, can still look based on this prefix (default
                        "node_info.json")
  -c CFSTATS_PREFIX, --cfstats_prefix CFSTATS_PREFIX
                        if cfstats in the diag tarball has an oddball name, can still look based on this prefix (default "cfstats")
  -b BLOCK_DEV_PREFIX, --block_dev_prefix BLOCK_DEV_PREFIX
                        if blockdev_report in the diag tarball has an oddball name, can still look based on this prefix (default
                        "blockdev_report")

Commands:
  {core,search,sysbottle,ttop}
    core                Cassandra and DSE Core specific sub-commands
    search              Search specific sub-commands
    sysbottle           sysbottle provides analysis of an iostat file. Supports iostat files generated via 'iostat -x -c -d -t'
    ttop                Analyze ttop files
`
}

//main is the entry point to the application
func main() {
	//subCmdArgs := os.Args[2:]
	subCmd := os.Args[1]
	switch subCmd {
	case "core":
	case "search":
	case "sysbottle":
	case "ttop":
	default:
		fmt.Printf("Unknown cmd %s\n", subCmd)
		fmt.Println(help())
	}
}
