package di

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type configurator func(ctx context.Context)

type constructor func(ctx context.Context)

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

type Application interface {
	Init(ctx context.Context)
	Done(ctx context.Context)
	Run(ctx context.Context)
}

func newApp() *App {
	return &App{
		dictionary: make(map[string]*component),
		variables:  NewVariables(nil),
	}
}

type App struct {
	mx         sync.Mutex
	components components
	dictionary map[string]*component
	variables  Variables
}

func (a *App) Variables() Variables {
	return a.variables
}

func (a *App) addComponent(component *component) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.components = append(a.components, component)
	a.dictionary[component.id] = component
}

func (a *App) Init(ctx context.Context) {
	a.sortComponents()
	for _, c := range a.components {
		if err := c.runInit(ctx); err != nil {
			panic(&componentError{c.name, fmt.Sprintf("init: %s", err.Error())})
		}
	}
}

func (a *App) Done(ctx context.Context) {
	ctx = context.WithValue(ctx, ApplicationContextKey, a)
	cs := a.components
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		c.runDone(ctx)
	}
}

func (a *App) Run(ctx context.Context) {
	// nothing to do
}

func (a *App) sortComponents() {
	a.mx.Lock()
	defer a.mx.Unlock()
	sort.Sort(&a.components)
}

func (a *App) get(ctx context.Context, name string, builder func(ctx context.Context) *component) *component {
	c := a.fetch(ctx, name)
	if c != nil {
		return c
	}

	c = builder(ctx)
	a.addComponent(c)

	return c
}

func (a *App) fetch(ctx context.Context, name string) *component {
	a.mx.Lock()
	defer a.mx.Unlock()

	c, _ := a.dictionary[name]
	return c
}

type AppOptions struct {
	configurators []configurator
	constructors  []constructor
}

type AppOption func(opts *AppOptions)

func newConstructor[T any](constructor Constructor[T]) func(ctx context.Context) {
	return func(ctx context.Context) {
		_ = constructor(ctx)
	}
}

func WithConfigurator(configurators ...func(ctx context.Context)) AppOption {
	return func(opts *AppOptions) {
		for _, c := range configurators {
			opts.configurators = append(opts.configurators, c)
		}
	}
}

func WithService[T any](constructor ...Constructor[T]) AppOption {
	return func(opts *AppOptions) {
		for _, c := range constructor {
			opts.constructors = append(opts.constructors, newConstructor(c))
		}
	}
}

func WithDaemon(daemons ...func(ctx context.Context)) AppOption {
	return func(opts *AppOptions) {
		for _, daemon := range daemons {
			opts.constructors = append(opts.constructors, daemon)
		}
	}
}

func Execute(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) {
	app, ctx, err := Build(ctx, constructor, options...)
	if err != nil {
		if e, ok := err.(*componentError); ok {
			Logger.Log(LogLevelError, e.component, e.message)
		} else {
			Logger.Log(LogLevelError, "application", err.Error())
		}
		return
	}
	defer app.Done(ctx)

	app.Run(ctx)
}

func Build(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) (a Application, c context.Context, err error) {
	app := newApp()
	ctx = context.WithValue(ctx, ApplicationContextKey, app)

	defer func() {
		if e := recover(); e != nil {
			err, _ = e.(error)
		}
	}()

	application := build(ctx, constructor, options...)
	application.Init(ctx)

	return application, ctx, err
}

func build(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) Application {
	var opts AppOptions
	for _, o := range options {
		o(&opts)
	}

	for _, c := range opts.configurators {
		c(ctx)
	}

	application := constructor(ctx)

	for _, c := range opts.constructors {
		c(ctx)
	}

	return application
}

type ApplicationContextType int

var ApplicationContextKey ApplicationContextType = 0

func GetAppFromContext(ctx context.Context) *App {
	app, _ := ctx.Value(ApplicationContextKey).(*App)
	if app == nil {
		panic(fmt.Errorf("application not found in context"))
	}
	return app
}
