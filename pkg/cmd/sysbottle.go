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
	"flag"
	"fmt"
	"os"
)

//ExecSysBottle is the entry point for the sysbottle subcommand
func ExecSysBottle(args []string) {
	sysBottleFlags := flag.NewFlagSet("sysbottle", flag.ExitOnError)
	cpuFlag := sysBottleFlags.Int("c", 50, "percentage cpu usage minus iowait% above this threshold is busy. Assumes hyperthreading is enabled (default 50)")
	diskQueueFlag := sysBottleFlags.Float64("q", 1.0, "disk queue depth. above this threshold is considered busy. (default 1)")
	disksFlag := sysBottleFlags.String("d", "", "comma separated list of disks to include in report, if no disks are provided then all found disks are included in report")
	ioWaitFlag := sysBottleFlags.Int("i", 5, "percentage iowait above this threshold is marked as a busy disk (default 5)")
	thresholdFlag := sysBottleFlags.Int("t", 5, "percentage of total time where we consider a node 'busy' for bottleneck summary. Example by default (5.0%) if the CPU and Disk are busy 5.0% of the total time measured then it is considered busy. (default 5)")
	if sysBottleFlags.NArg() != 1 {
		sysBottleFlags.Usage()
		os.Exit(2)
	}
	file := sysBottleFlags.Arg(1)

	if err := sysBottleFlags.Parse(args); err != nil {
		fmt.Printf("error parsing flags %s\n", err)
		os.Exit(2)
	}
	if report, err := RunSysBottle(*cpuFlag, *diskQueueFlag, *disksFlag, *ioWaitFlag, *thresholdFlag, file); err != nil {
		fmt.Printf("sysbottle error %s\n", err)
		os.Exit(2)
	} else {
		fmt.Println(report)
	}
}

//RunSysBottle parses the iostat file and generates a report based on the command line flags
func RunSysBottle(cpu int, diskQueue float64, disks string, ioWait int, threshold int, file string) (string, error) {
	var report string
	return report, nil
}

/**
func parseFiles(){
    # states
    cpu := 'cpu'
    date := 'date'
    device := 'device'

    #us, eu date formats, and one seen in sper66
    datefmts := ['%m/%d/%Y %I:%M:%S %p', '%d/%m/%y %H:%M:%S', '%m/%d/%y %H:%M:%S']

        .state = None
        self.__mkiostat()

    def __mkiostat(self):
        self.iostat = {
            'cpu': {'cols': [], 'stat': []},
            'device': {'cols': [], 'stat': {}},
            'date': None
            }

    def _parse(self, line):
        if line == '\n': # empty lines are the reset switch
            if self.state == self.DEVICE:
                yield self.iostat
                self.__mkiostat()
            self.state = None
            return

        line = line.strip()

        if self.state == self.CPU:
            self._parse_cpu(line)
        elif self.state == self.DEVICE:
            self._parse_device(line)
        elif line[0].isdigit():
            self._parse_date(line)
        else:
            if line.startswith('avg-cpu'):
                self.state = self.CPU
                self.iostat['cpu']['cols'] = self._parse_columns(line)
            elif line.startswith('Device'):
                self.state = self.DEVICE
                self.iostat['device']['cols'] = self._parse_columns(line)

    def _parse_columns(self, line):
        return line.split()[1:]

    def _parse_cpu(self, line):
        self.iostat['cpu']['stat'] = [float(i.replace(',', '.')) for i in line.split()]

    def _parse_device(self, line):
        parts = line.split()
        self.iostat['device']['stat'][parts[0]] = [float(i.replace(',', '.')) for i in parts[1:]]

    def _parse_date(self, line):
        date = None
        for datefmt in self.datefmts:
            try:
                date = datetime.strptime(line, datefmt)
            except ValueError as e:
                if env.DEBUG:
                    print(e)
        if not date:
            raise ValueError("tried parsing in the following formats " + \
                    "'%s' but %s does not match" % (self.datefmts, line))
        self.iostat['date'] = date

    def parse(self, infile):
        "parse an iostat file"
        with open(infile, 'r') as f:
            for line in f:
                for stat in self._parse(line):
                    yield stat

*/
