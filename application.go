package di

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

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
}

func newApp() *App {
	return &App{}
}

type App struct {
	mx         sync.Mutex
	components components
}

func (a *App) addComponent(component *component) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.components = append(a.components, component)
}

func (a *App) Init(ctx context.Context) {
	a.sortComponents()
	for _, c := range a.components {
		if err := c.runInit(ctx); err != nil {
			panic(fmt.Errorf("%s: %w", c.name, err))
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

func (a *App) sortComponents() {
	a.mx.Lock()
	defer a.mx.Unlock()
	sort.Sort(&a.components)
}

type AppOptions struct {
	constructors []constructor
}

type AppOption func(opts *AppOptions)

func newConstructor[T any](constructor Constructor[T]) func(ctx context.Context) {
	return func(ctx context.Context) {
		_ = constructor(ctx)
	}
}

func WithWorker[T any](constructor ...Constructor[T]) AppOption {
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

func Build[T Application](
	ctx context.Context,
	constructor Constructor[T],
	options ...AppOption,
) (application T, err error) {
	var opts AppOptions
	for _, o := range options {
		o(&opts)
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	app := newApp()
	ctx = context.WithValue(ctx, ApplicationContextKey, app)

	application = constructor(ctx)

	for _, c := range opts.constructors {
		c(ctx)
	}

	application.Init(ctx)
	return application, nil
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
