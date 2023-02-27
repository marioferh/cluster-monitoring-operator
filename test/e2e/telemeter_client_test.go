// Copyright 2023 The Cluster Monitoring Operator Authors
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

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift/cluster-monitoring-operator/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestTelemeterClient verifies that the telemeter client can collect metrics from the monitoring stack and forward them to the telemeter server.

func TestTelemeterClient(t *testing.T) {
	{
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterMonitorConfigMapName,
				Namespace: f.Ns,
				Labels: map[string]string{
					framework.E2eTestLabelName: framework.E2eTestLabelValue,
				},
			},
			Data: map[string]string{
				"config.yaml": "{}",
			},
		}
		f.MustCreateOrUpdateConfigMap(t, cm)
		t.Cleanup(func() {
			f.MustDeleteConfigMap(t, cm)
		})

		ctx := context.Background()
		patch := []byte(`{
	"spec": {
		"overrides": [{
			"group": "apps",
			"kind": "Deployment",
			"name": "cluster-monitoring-operator",
			"namespace": "openshift-monitoring",
		}]
	}
}`)
		_, err := f.OpenShiftConfigClient.ConfigV1().ClusterVersions().Patch(ctx, "version", types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			patch := []byte(`{"spec": {"overrides": []}}`)
			_, _ = f.OpenShiftConfigClient.ConfigV1().ClusterVersions().Patch(ctx, "version", types.MergePatchType, patch, metav1.PatchOptions{})
		})

		f.PrometheusK8sClient.WaitForQueryReturn(
			t,
			5*time.Minute,
			`metricsclient_request_send{client="federate_to",job="telemeter-client",status_code="200"}`,
			func(v float64) error {
				if v == 0 {
					return fmt.Errorf("expecting metricsclient request send more than 0 but got none")
				}
				return nil
			},
		)

		f.PrometheusK8sClient.WaitForQueryReturn(
			t,
			5*time.Minute,
			`federate_samples{job="telemeter-client"}`,
			func(v float64) error {
				if v < 10 {
					return fmt.Errorf("expecting federate samples from telemeter client more than 10 but got %f", v)
				}
				return nil
			},
		)
	}
}
