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
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/DataStax-Toolkit/sperf/pkg/env"
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

type CPUMeasurement struct {
	Cols []string
	Stat []float64
}

type DeviceMeasurement struct {
	Cols []string
	Stat map[string][]float64
}

type IOStatRow struct {
	Date   time.Time
	CPU    CPUMeasurement
	Device DeviceMeasurement
}

//IOStatParser reads an iostat file and pulls out the numbers we care about
type IOStatParser struct {
	state  string
	iostat IOStatRow
}

func (s *IOStatParser) mkiostat() {
	s.iostat = IOStatRow{
		CPU: CPUMeasurement{},
		Device: DeviceMeasurement{
			Stat: make(map[string][]float64),
		},
	}
}

// states
var CPU = "cpu"
var DATE = "date"
var DEVICE = "device"

//results := make(chan IOStatRow)
//us, eu date formats, and one seen in sper66

var datefmts []string = []string{"%m/%d/%Y %I:%M:%S %p", "%d/%m/%y %H:%M:%S", "%m/%d/%y %H:%M:%S"}

func (s *IOStatParser) ParseRow(results chan IOStatRow, line string) error {
	if line == "\n" { // empty lines are the reset switch
		if s.state == DEVICE {
			results <- s.iostat
			s.mkiostat()
		}
		s.state = ""
		return nil
	}

	line = strings.TrimSpace(line)

	if s.state == CPU {
		if err := s.parseCPU(line); err != nil {
			return fmt.Errorf("unable to parse cpu with error %s", err)
		}
	} else if s.state == DEVICE {
		if err := s.parseDevice(line); err != nil {
			return fmt.Errorf("unable to parse device with error %s", err)
		}
	} else if unicode.IsDigit(rune(line[0])) {
		if err := s.parseDate(line); err != nil {
			return fmt.Errorf("unable to parse date with error %s", err)
		}
	} else {
		if strings.HasPrefix(line, "avg-cpu") {
			s.state = CPU
			s.iostat.CPU.Cols = s.parseColumns(line)
		} else if strings.HasPrefix(line, "Device") {
			s.state = DEVICE
			s.iostat.Device.Cols = s.parseColumns(line)
		}
	}
	return nil
}

func (s *IOStatParser) parseColumns(line string) []string {
	return strings.Split(line, " ")
}

func (s *IOStatParser) parseCPU(line string) error {
	var ioStats []float64
	for _, rawStat := range strings.Split(line, " ") {
		stat := strings.Replace(rawStat, ",", ".", 1)
		parsedStat, err := strconv.ParseFloat(stat, 64)
		if err != nil {
			return fmt.Errorf("unable to parse cpu line '%s' for number '%s' with error '%s'", line, stat, err)
		}
		ioStats = append(ioStats, parsedStat)
	}
	s.iostat.CPU.Stat = ioStats
	return nil
}

func (s *IOStatParser) parseDevice(line string) error {
	parts := strings.Split(line, " ")
	var deviceStat []float64
	for _, rawStat := range parts[1:] {
		stat := strings.Replace(rawStat, ",", ".", 1)
		parsedStat, err := strconv.ParseFloat(stat, 64)
		if err != nil {
			return fmt.Errorf("unable to parse device line '%s' for number '%s' with error '%s'", line, stat, err)
		}
		deviceStat = append(deviceStat, parsedStat)
	}
	deviceName := parts[0]
	s.iostat.Device.Stat[deviceName] = deviceStat
	return nil
}

func (s *IOStatParser) parseDate(line string) error {
	var date time.Time
	for _, datefmt := range datefmts {
		if rawDate, err := time.Parse(line, datefmt); err != nil {
			if env.Debug {
				fmt.Printf("unable to parse format %s with line %s", datefmt, line)
			}
		} else {
			date = rawDate
		}
	}
	if date.IsZero() {
		return fmt.Errorf("tried parsing in the following formats '%s' but %s does not match", datefmts, line)
	}
	s.iostat.Date = date
	return nil
}

func (s *IOStatParser) Parse(infile string) (chan IOStatRow, error) {
	//parse an iostat file
	results := make(chan IOStatRow)
	f, err := os.Open(infile)
	if err != nil {
		return results, fmt.Errorf("unable to open file %s with error %s", infile, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := s.ParseRow(results, scanner.Text()); err != nil {
			return results, fmt.Errorf("unable to parse row with error %s", err)
		}
	}
	return results, nil
}
