package probes

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/operator"
	"github.com/grafana/agent/component/prometheus/operator/common"
	"github.com/grafana/agent/service/cluster"
	"github.com/grafana/agent/service/http"
	"github.com/grafana/agent/service/labelstore"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.operator.probes",
		Args:          operator.Arguments{},
		NeedsServices: []string{cluster.ServiceName, http.ServiceName, labelstore.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return common.New(opts, args, common.KindProbe)
		},
	})
}
