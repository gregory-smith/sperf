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
package cmd

import (
	//	"flag"
	"fmt"
	"os"

	"github.com/DataStax-Toolkit/sperf/pkg/bgrep"
	"github.com/DataStax-Toolkit/sperf/pkg/diag"
	"github.com/DataStax-Toolkit/sperf/pkg/gc"
	"github.com/DataStax-Toolkit/sperf/pkg/jarcheck"
	"github.com/DataStax-Toolkit/sperf/pkg/slowquery"
	"github.com/DataStax-Toolkit/sperf/pkg/statuslogger"
)

//ExecCore is the entry point for the core subcommand
func ExecCore(args []string) {
	helpText := `usage: sperf core [-h]
                  {bgrep,diag,gc,jarcheck,schema,slowquery,statuslogger} ...

optional arguments:
  -h, --help            show this help message and exit

DSE Core/Cassandra Commands:
  {bgrep,diag,gc,jarcheck,schema,slowquery,statuslogger}
    bgrep               search for custom regex and bucketize results
    diag                Generates a diagtarball report. DSE 5.0-6.7
    gc                  show gc info. provides time series of gc duration and
                        frequency
    jarcheck            Checks jar versions in output.logs. Supports tarballs
                        and files. DSE 5.0-6.7
    schema              Analyze schema for summary. DSE 5.0-6.7
    slowquery           Generates a report of slow queries in debug log. DSE
                        6.0-6.7. DEPRECATED use 'sperf core slowquery' instead
    statuslogger        Provides analysis of StatusLogger log lines. DSE
                        5.0-6.7

`
	if len(args) == 0 {
		fmt.Println(helpText)
		os.Exit(0)
	}
	subCmd := args[1]
	switch subCmd {
	case "bgrep":
		bgrep.Exec(args)
	case "diag":
		diag.Exec(args)
	case "gc":
		gc.Exec(args)
	case "jarcheck":
		jarcheck.Exec(args)
	case "slowquery":
		slowquery.Exec(args)
	case "statuslogger":
		statuslogger.Exec(args)
	default:
		fmt.Printf("subcommand %s is not supported\n", subCmd)
		os.Exit(1)
	}
}
