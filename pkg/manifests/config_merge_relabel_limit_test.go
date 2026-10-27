// Copyright 2026 The Cluster Monitoring Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifests

import (
	"fmt"
	"testing"

	configv1alpha1 "github.com/openshift/api/config/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestConfig_RemoteWriteRelabelRuleCount(t *testing.T) {
	for _, count := range []int{0, 1, 10, 11, 100} {
		t.Run(fmt.Sprintf("%d rules", count), func(t *testing.T) {
			rules := make([]configv1alpha1.RelabelConfig, count)
			for i := range rules {
				rules[i] = configv1alpha1.RelabelConfig{
					// Reverse names to detect accidental sorting of the ordered rules.
					Name:         fmt.Sprintf("rule-%03d", count-i),
					SourceLabels: []string{"job"},
					Regex:        fmt.Sprintf("^metric_%d$", i),
					Action: configv1alpha1.RelabelActionConfig{
						Type: configv1alpha1.RelabelActionKeep,
					},
				}
			}
			cr := &configv1alpha1.ClusterMonitoring{
				Spec: configv1alpha1.ClusterMonitoringSpec{
					PrometheusConfig: configv1alpha1.PrometheusConfig{
						RemoteWrite: []configv1alpha1.RemoteWriteSpec{{
							Name: "rw", URL: "https://remote.example.com/write",
							WriteRelabelConfigs: rules,
						}},
					},
				},
			}
			merged, err := NewConfigFromStringAndClusterMonitoringResource("{}", cr)
			require.NoError(t, err)
			remoteWrite := merged.ClusterMonitoringConfiguration.PrometheusK8sConfig.RemoteWrite
			require.Len(t, remoteWrite, 1)
			actual := remoteWrite[0].WriteRelabelConfigs
			require.Len(t, actual, count)
			for i, expected := range rules {
				require.Equal(t, expected.Regex, actual[i].Regex, "rule %d", i)
				require.Equal(t, string(expected.Action.Type), actual[i].Action, "rule %d", i)
				require.Len(t, actual[i].SourceLabels, 1)
				require.EqualValues(t, "job", actual[i].SourceLabels[0], "rule %d", i)
			}
		})
	}
}
