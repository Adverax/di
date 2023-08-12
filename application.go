package di

import (
	"fmt"
	"sort"
	"sync"
)

type State int

const (
	StateBuild State = iota
	StateInit
	StateDone
)

type component struct {
	state    State
	init     func() error
	done     func()
	priority int
	name     string
}

func (c *component) runInit() error {
	if c.state != StateBuild {
		return nil
	}
	c.state = StateInit
	if c.init == nil {
		return nil
	}
	err := c.init()
	if err != nil {
		return fmt.Errorf("initialization error %w", err)
	}
	Logger.Log(fmt.Sprintf("%s: initialized", c.name))
	return nil
}

func (c *component) runDone() {
	if c.state != StateInit {
		return
	}
	c.state = StateDone
	if c.done == nil {
		return
	}
	c.done()
	Logger.Log(fmt.Sprintf("%s: done", c.name))
}

type components []*component

func (c components) Len() int {
	return len(c)
}

func (c components) Less(i, j int) bool {
	return c[i].priority < c[j].priority
}

func (c components) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type app struct {
	mx         sync.Mutex
	components components
}

func (a *app) addComponent(component *component) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.components = append(a.components, component)
}

func (a *app) Init() {
	a.mx.Lock()
	defer a.mx.Unlock()

	sort.Sort(&a.components)

	for _, c := range a.components {
		if err := c.runInit(); err != nil {
			panic(fmt.Errorf("%s: %w", c.name, err))
		}
	}
}

func (a *app) Done() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for i := len(a.components) - 1; i >= 0; i-- {
		c := a.components[i]
		c.runDone()
	}
}

type App interface {
	Init()
	Done()
}

var application = &app{}

func GetApp() App {
	return application
}

type AppOptions struct {
	workers []func()
}

type AppOption func(opts *AppOptions)

func newBuilder[T any](builder func() T) func() {
	return func() {
		_ = builder()
	}
}

func WithWorker[T any](builder ...func() T) AppOption {
	return func(opts *AppOptions) {
		for _, b := range builder {
			opts.workers = append(opts.workers, newBuilder(b))
		}
	}
}

func WithDaemon(daemon ...func()) AppOption {
	return func(opts *AppOptions) {
		opts.workers = append(opts.workers, daemon...)
	}
}

func Build[T App](
	builder func() T,
	options ...AppOption,
) (app T, err error) {
	var opts AppOptions
	for _, o := range options {
		o(&opts)
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	app = builder()

	for _, worker := range opts.workers {
		worker()
	}

	app.Init()
	return app, nil
}
