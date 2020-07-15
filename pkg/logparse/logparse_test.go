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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLog(t *testing.T) {
	counter := 0
	reader := bytes.NewBufferString("a\nb\nc\n")
	readLine := func(s string) (map[string]interface{}, error) {
		result := make(map[string]interface{})
		counter += 1
		result[s] = counter
		return result, nil
	}
	resultsChan := ReadLog(reader, readLine, nil)
	first := <-resultsChan

	assert.Nil(t, first.Err)
	assert.Equal(t, map[string]interface{}{
		"a": 1,
	}, first.Row)
	second := <-resultsChan
	assert.Nil(t, second.Err)
	assert.Equal(t, map[string]interface{}{
		"b": 2,
	}, second.Row)
	third := <-resultsChan
	assert.Nil(t, third.Err)
	assert.Equal(t, map[string]interface{}{
		"c": 3,
	}, third.Row)
}
