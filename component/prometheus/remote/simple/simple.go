package simple

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/remote"
	"github.com/prometheus/prometheus/storage"
)

func init() {
	component.Register(component.Registration{
		Name:      "prometheus.remote.simple",
		Singleton: false,
		Args:      Arguments{},
		Exports:   Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

func NewComponent(opts component.Options, args Arguments) (*Simple, error) {
	database, err := newDBStore(false, args.TTL, 5*time.Minute, path.Join(opts.DataPath, "wal"), opts.Logger)
	if err != nil {
		return nil, err
	}
	s := &Simple{
		database: database,
		opts:     opts,
	}
	return s, s.Update(args)
}

type Simple struct {
	mut      sync.RWMutex
	database *dbstore
	args     Arguments
	opts     component.Options
}

// Run starts the component, blocking until ctx is canceled or the component
// suffers a fatal error. Run is guaranteed to be called exactly once per
// Component.
//
// Implementations of Component should perform any necessary cleanup before
// returning from Run.
func (s *Simple) Run(ctx context.Context) error {
	ctx.Done()
	return nil
}

// Update provides a new Config to the component. The type of newConfig will
// always match the struct type which the component registers.
//
// Update will be called concurrently with Run. The component must be able to
// gracefully handle updating its config while still running.
//
// An error may be returned if the provided config is invalid.
func (s *Simple) Update(args component.Arguments) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.args = args.(Arguments)
	s.opts.OnStateChange(Exports{Receiver: s})

	return nil
}

// Appender returns a new appender for the storage. The implementation
// can choose whether or not to use the context, for deadlines or to check
// for errors.
func (c *Simple) Appender(ctx context.Context) storage.Appender {
	return newAppender(c)
}

func (c *Simple) commit(a *appender) {
	c.mut.RLock()
	defer c.mut.Unlock()

	endpoint := time.Now().UnixMilli() - int64(c.args.TTL.Seconds())

	timestampedMetrics := make([]any, 0)
	for _, x := range a.metrics {
		// No need to write if already outside of range.
		if x.Timestamp < endpoint {
			continue
		}
		timestampedMetrics = append(timestampedMetrics, x)
	}

	c.database.WriteSignal(timestampedMetrics)
}

type Arguments struct {
	TTL time.Duration `river:"ttl,attr,optional"`

	Endpoint *remote.EndpointOptions `river:"endpoint,block,optional"`
}

type Exports struct {
	Receiver storage.Appendable `river:"receiver,attr"`
}
