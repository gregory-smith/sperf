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
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"
)

//main is the entry point to a script written in go that creates a zip file
//this is only used by the deploy task in the Makefile and is used for ad-hoc builds
//I wanted a way to have a cross platform zip without installing an additional tool
//since this was short and easy and I borrowed it from a blog post on the interet
//this seemed like minimal effort to use.
func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: go run zip.go myzip.zip \"myfile1, myfile2, myfile3\"\n\tzip command requires two args filename and a comma separated list of files")
		fmt.Printf("zip command requires two args filename and a comma separated list of files")
		fmt.Printf("\norgs were parsed as %s\n", os.Args)
		os.Exit(1)
	}
	filename := os.Args[1]
	//split on comma
	filesRaw := strings.Split(os.Args[2], ",")
	var files []string
	//strip all whitespace
	for _, file := range filesRaw {
		files = append(files, strings.TrimSpace(file))
	}
	//create the zip file
	newZipFile, err := os.Create(filename)
	if err != nil {
		//throw an error and log the reason for failure
		fmt.Printf("unexpected error creating zip: %s\n", err)
		os.Exit(1)
	}
	//make sure we close the file
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	//make sure the zip writer is closed even if the program exits
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			fmt.Printf("unexpected error adding file to zip: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s added to zip file\n", file)
	}
	fmt.Printf("zip %s successfully created\n\n", filename)
}

//AddFileToZip adds individual files to an open zip file
func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	//need to open up the file to add it
	fileToZip, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("unable to open file %s with error %s", filename, err)
	}
	defer fileToZip.Close()

	// Get the file information so we can read the header
	info, err := fileToZip.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file information for file %s with error %s", filename, err)
	}

	//need to get this to change the algorithm and the full path
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("unable to get header info for file %s with error %s", filename, err)
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	//add the file header to the zip file so it can know about it
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("unable to create header for zip file: %s", err)
	}
	//copy to file to the writer and now it should be added
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return fmt.Errorf("unable to copy new file %s to the zip file with error %s", filename, err)
	}
	return nil
}
