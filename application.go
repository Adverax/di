package di

import (
	"sync"
)

type component struct {
	init func() error
	done func()
}

func (c *component) runInit() error {
	if c.init == nil {
		return nil
	}
	return c.init()
}

func (c *component) runDone() {
	if c.done == nil {
		return
	}
	c.done()
}

type app struct {
	mx         sync.Mutex
	components []*component
}

func (a *app) addComponent(component *component) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.components = append(a.components, component)
}

func (a *app) Init() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, c := range a.components {
		if err := c.runInit(); err != nil {
			panic(err)
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

func WithWorker[T any](builder func() T) AppOption {
	return func(opts *AppOptions) {
		opts.workers = append(
			opts.workers,
			func() {
				_ = builder()
			},
		)
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
