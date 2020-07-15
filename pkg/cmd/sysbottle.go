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
	reporter := NewSysBottleReport(file, &SysBottleConf{})
	return "", reporter.PrintReport()
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
var datefmts []string = []string{
	"01/02/2006 03:04:05 PM",
	"02/01/2006 15:04:05",
	"01/02/2006 15:04:05",
}

func (s *IOStatParser) parseRow(results chan IOStatRow, line string) error {
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

func (s *IOStatParser) parseFile(infile string) (chan IOStatRow, chan<- error) {
	//parse an iostat file
	results := make(chan IOStatRow)
	errCh := make(chan error)
	f, err := os.Open(infile)
	if err != nil {
		errCh <- fmt.Errorf("unable to open file %s with error %s", infile, err)
		return results, errCh
	}
	defer f.Close()
	go func() {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if err := s.parseRow(results, scanner.Text()); err != nil {
				errCh <- fmt.Errorf("unable to parse row with error %s", err)
				return
			}
		}
	}()
	return results, errCh
}

type SysBottleConf struct {
	IOWaitThreshold int
	CPUThreshold    int
	Disks           []string
	QueueThreshold  int
	BusyThreshold   int
}

//Produces a report from iostat output"
type SysBottleReport struct {
	infile         string
	parser         IOStatParser
	count          int
	cpuExceeded    int
	ioWaitExceeded int
	cpuStats       map[string][]float64
	queueDepth     map[string]int
	devices        map[string]func() map[string][]string
	start          time.Time
	end            time.Time
	deviceIndex    map[string]int
	cpuIndex       map[string]int
	conf           SysBottleConf
	recs           map[string]bool
	analyzed       bool
}

func NewSysBottleReport(infile string, conf *SysBottleConf) *SysBottleReport {
	s := SysBottleReport{}
	s.infile = infile
	s.parser = IOStatParser{}

	s.count = 0
	s.cpuExceeded = 0
	s.ioWaitExceeded = 0
	s.devices = make(map[string]func() map[string][]string)
	s.cpuStats = make(map[string][]float64)
	s.queueDepth = make(map[string]int)
	s.deviceIndex = make(map[string]int)
	s.cpuIndex = make(map[string]int)
	if conf == nil {
		s.mkConf()
	} else {
		s.conf = *conf
	}
	s.recs = make(map[string]bool)
	s.analyzed = false
	return &s
}

func (s *SysBottleReport) mkConf() SysBottleConf {
	return SysBottleConf{
		IOWaitThreshold: 5,
		CPUThreshold:    50,
		Disks:           []string{},
		QueueThreshold:  1,
		BusyThreshold:   5,
	}
}

func (s *SysBottleReport) analyze() error {
	//analyzes the file this class was initialized with
	ch, err := s.parser.parseFile(s.infile)
	if err != nil {
		return fmt.Errorf("unable to analyze the file with error: %s", err)
	}

	for row := range ch {
		s.count += 1
		if s.deviceIndex != nil {
			s.mkColIDX(row)
		}
		s.analyzeDisk(row)
		s.analyzeCPU(row)
		if !s.start.IsZero() {
			s.start = row.Date
		}
		if !s.end.IsZero() || row.Date.After(s.end) {
			s.end = row.Date
		}
		s.analyzed = true
	}
	return nil
}

func (s *SysBottleReport) mkColIDX(stat IOStatRow) {
	for i, col := range stat.Device.Cols {
		s.deviceIndex[col] = i
	}
	for i, col := range stat.CPU.Cols {
		s.cpuIndex[col] = i
	}
}

func (s *SysBottleReport) wantDisk(name string) bool {
	if len(s.conf.Disks) == 0 {
		//this is not super obvious but the "no disks" case means all disks match
		return true
	}
	for _, diskName := range s.conf.Disks {
		if diskName == name {
			return true
		}
	}
	return false
}

func (s *SysBottleReport) analyzeDisk(stat IOStatRow) {
	for disk, values := range stat.Device.Stat {
		if s.wantDisk(disk) {
			for col := range s.deviceIndex {
				val := values[s.deviceIndex[col]]
				s.devices[disk][col] = append(s.devices[disk][col], val)
				//TODO fix this int casting to val this was a bug in the old python code that was comparing ints and floats
				if strings.Contains(col, "qu") && int(val) >= s.conf.QueueThreshold {
					s.queueDepth[disk] += 1
					s.recs[fmt.Sprintf("* decrease activity on %s", disk)] = true
				}
			}
		}
	}
}

func (s *SysBottleReport) analyzeCPU(stat IOStatRow) {
	total := 0
	for _, cpu := range []string{"system", "user", "nice", "steal"} {
		//TODO fix this casting, this was a bug in the old python codebase
		//that never got caught
		total += int(stat.CPU.Stat[s.cpuIndex["%"+cpu]])
	}
	//TODO fix this casting, this was a bug in the old python codebase
	//that never got caught
	s.cpuStats["total"] = append(s.cpuStats["total"], float64(total))
	if total > s.conf.CPUThreshold {
		s.cpuExceeded += 1
		s.recs["* tune for less CPU usage"] = true
	}
	for col := range s.cpuIndex {
		val := stat.CPU.Stat[s.cpuIndex[col]]
		s.cpuStats[col] = append(s.cpuStats[col], val)
	}
	ioWaitIndex := s.cpuIndex["%iowait"]
	//TODO fix this casting, this was a bug in the old python codebase
	//that never got caught
	if int(stat.CPU.Stat[ioWaitIndex]) > s.conf.IOWaitThreshold {
		s.ioWaitExceeded += 1
		s.recs["* tune for less IO"] = true
	}
}

//PrintReport prints a report for the file this class was initialized with, analyzing if necessary"
func (s *SysBottleReport) PrintReport() error {
	if !s.analyzed {
		if err := s.analyze(); err != nil {
			return fmt.Errorf("unable to generate analysis %s", err)
		}
	}
	fmt.Println("sysbottle\n")
	fmt.Printf("* total records: %s\n", s.count)
	if s.count > 0 {
		reportPercentage :=
			func(a int) float64 {
				return (float64(a) / float64(s.count)) * 100.0
			}
		fmt.Printf("* total bottleneck time: %.2f%% (cpu bound, io bound, or both)\n",
			reportPercentage(s.ioWaitExceeded+s.cpuExceeded))
		fmt.Printf("* cpu+system+nice+steal time > %.2f%%: %.2f%%\n",
			s.conf.CPUThreshold, reportPercentage(s.cpuExceeded))
		fmt.Printf("* iowait time > %.2f%%: %.2f%%\n",
			s.conf.IOWaitThreshold, reportPercentage(s.ioWaitExceeded))
		fmt.Printf("* start %s\n", s.start)
		fmt.Printf("* end %s\n", s.end)
		logTimeSeconds := s.end.Second() - s.start.Second() + 1
		fmt.Printf("* log time: %ss\n", logTimeSeconds)
		fmt.Printf("* interval: %ss\n", reportPercentage(logTimeSeconds))
		for device := range s.devices {
			fmt.Printf("* %s time at queue depth >= %.2f: %.2f%%\n",
				device, s.conf.QueueThreshold, reportPercentage(s.queueDepth[device]))
		}
		fmt.Println("")
		var lines [][]string
		lines = append(lines, getPercentileHeaders())
		lines = append(lines, []string{"", "---", "---", "---", "---", "---", "---"})
		lines = append(lines, getPercentiles("cpu", s.cpuStats["total"]))
		lines = append(lines, getPercentiles("iowait", s.cpuStats["%iowait"]))
		lines = append(lines, []string{})
		lines = append(lines, getPercentileHeaders())
		lines = append(lines, []string{"", "---", "---", "---", "---", "---", "---"})
		for device := range s.devices {
			lines = append(lines, []string{device, "", "", "", "", "", ""})
			for iotype := range s.devices[device]() {
				if strings.Contains(iotype, "qu") || strings.Contains(iotype, "wait") {
					lines = append(lines, getPercentiles("- "+iotype+":", s.devices[device][iotype]))
				}
			}
		}
		lines = append(lines, []string{})
		humanize.padTable(lines, 8, 2)
		for _, line := range lines {
			fmt.Println(strings.Join(line, ""))
		}
		s.printRecommendations()
	}
}

func (s *SysBottleReport) printRecommendations() {
	if len(s.recs) == 0 {
		return
	}
	fmt.Println("recommendations")
	fmt.Println(strings.Repeat("-", 15))
	for _, rec := range s.recs {
		fmt.Println(rec)
	}
}
