package di

import (
	"errors"
	"reflect"
	"sync"
)

type Phase string

const (
	PhaseConstruct Phase = "construct"
	PhaseInit      Phase = "init"
	PhaseDone      Phase = "done"
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
	phase      Phase
	components []*component
	workers    []func()
}

func (a *app) addWorker(worker func()) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.workers = append(a.workers, worker)
}

func (a *app) addComponent(component *component) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.components = append(a.components, component)
}

func (a *app) Init() {
	a.phase = PhaseInit

	for _, worker := range a.workers {
		worker()
	}

	a.mx.Lock()
	defer a.mx.Unlock()

	for _, c := range a.components {
		if err := c.runInit(); err != nil {
			panic(err)
		}
	}
}

func (a *app) Done() {
	a.phase = PhaseDone

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

var application = &app{phase: PhaseConstruct}

func GetApp() App {
	return application
}

type Option[T any] func(options *Options[T])

type Options[T any] struct {
	init   func(instance T) error
	done   func()
	worker bool
}

func WithInit[T any](initializer func(instance T) error) Option[T] {
	return func(options *Options[T]) {
		options.init = initializer
	}
}

func WithDone[T any](finalizer func()) Option[T] {
	return func(options *Options[T]) {
		options.done = finalizer
	}
}

func AsWorker[T any]() Option[T] {
	return func(options *Options[T]) {
		options.worker = true
	}
}

func NewComponent[T any](
	constructor func() T,
	options ...Option[T],
) func() T {
	var opts Options[T]
	for _, o := range options {
		o(&opts)
	}

	var instance T

	builder := func() T {
		if isZeroVal(instance) {
			if application.phase != PhaseConstruct && !opts.worker {
				panic(errors.New("component must be created in construct phase"))
			}
			instance = constructor()
			application.addComponent(newComponent(instance, opts))
		}

		return instance
	}

	if opts.worker {
		application.addWorker(func() {
			_ = builder()
		})
	}

	return builder
}

func newComponent[T any](instance T, opts Options[T]) *component {
	return &component{
		init: func() error {
			if opts.init != nil {
				return opts.init(instance)
			}
			return nil
		},
		done: func() {
			if opts.done != nil {
				opts.done()
			}
		},
	}
}

// IsZeroVal check if any type is its zero value
func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func Build[T App](builder func() T) (app T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	app = builder()
	app.Init()
	return app, nil
}
