// Copyright Project Harbor Authors
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

package logger

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRetrieve(t *testing.T) {
	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
		},
		{
			Name:  "FILE",
			Level: "INFO",
			Settings: map[string]interface{}{
				"base_dir": os.TempDir(),
			},
			Sweeper: &config.LogSweeperConfig{
				Duration: 1,
				Settings: map[string]interface{}{
					"work_dir": os.TempDir(),
				},
			},
		},
	}

	err := Init(context.TODO())
	require.NoError(t, err)
	_, err = Retrieve("no_id")
	require.Error(t, err)
}
