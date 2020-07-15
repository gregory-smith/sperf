// Copyright 2020 DataStax, Inc
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
package logparse

import (
	"bufio"
	"fmt"
	"io"

	"github.com/DataStax-Toolkit/sperf/pkg/logparse/blockdev"
	"github.com/DataStax-Toolkit/sperf/pkg/logparse/outputlog"
	"github.com/DataStax-Toolkit/sperf/pkg/logparse/systemlog"
)

//ReadSystemLog reads the system log, yields an iterable set of events of parsed logs
func ReadSystemLog(lines io.Reader, extras map[string]interface{}) <-chan ProcessedLine {
	return ReadLog(lines, systemlog.ReadLine, extras)
}

//ReadOutputLog reads the output log, yields an channel of events of parsed logs
func ReadOutputLog(lines io.Reader, extras map[string]interface{}) <-chan ProcessedLine {
	return ReadLog(lines, outputlog.ReadLine, extras)
}

//ReadBlockDev consumes a block dev report, yields channel of events of parsed logs
//just consume the channel to read the data
func ReadBlockDev(lines io.Reader, extras map[string]interface{}) <-chan ProcessedLine {
	return ReadLog(lines, blockdev.ReadLine, extras)
}

//ReadLine is the way we capture lines
type ReadLine = func(string) (map[string]interface{}, error)

//ProcessedLine is the result of each processed line in a text file
type ProcessedLine struct {
	Row map[string]interface{}
	Err error
}

//ReadLog runs a capture rule through and matches all lines
func ReadLog(r io.Reader, readLine ReadLine, extras map[string]interface{}) <-chan ProcessedLine {
	//this maps well to the python generator pattern we were using in logparse, keeping it around for now. Will
	//compare with and without channels later
	c := make(chan ProcessedLine)
	scanner := bufio.NewScanner(r)
	go func() {
		var fields map[string]interface{}
		//	make(map[string]interface{})
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				c <- ProcessedLine{
					Row: nil,
					Err: fmt.Errorf("unable to process read file with error %s", err),
				}
				//stop processing
				break
			}
			line := scanner.Text()
			nextFields, err := readLine(line)
			if err != nil {
				c <- ProcessedLine{
					Row: nil,
					Err: fmt.Errorf("unable to process line '%s' with error %s", line, err),
				}
				//stop processing
				break
			}
			if nextFields != nil {
				if fields != nil {
					for k, v := range extras {
						fields[k] = v
					}
					//successfully processed send it on
					c <- ProcessedLine{
						Row: fields,
						Err: nil,
					}
					//intentionally not exiting the loop here so we can set the fields with the just processed ones
				}
				//set this up for the next row NOTE: I don't remember why we do this..but there is a good reason. TODO research reason
				fields = nextFields
			}
		}
		//now that we're done scanning clear out the last processed row
		if fields != nil {
			c <- ProcessedLine{
				Row: fields,
				Err: nil,
			}
		}
	}()
	return c
}
