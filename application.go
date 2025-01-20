package di

import (
	"context"
	"fmt"
	"github.com/adverax/log"
	"sync"
)

type configurator func(ctx context.Context)

type constructor func(ctx context.Context)

type components []*component

// Application - interface for application
type Application interface {
	Init(ctx context.Context)
	Done(ctx context.Context)
	Run(ctx context.Context)
}

func newApp(logger log.Logger) *App {
	return &App{
		dictionary: make(map[string]*component),
		logger:     logger,
	}
}

// App - base implementation of Application
type App struct {
	mx         sync.Mutex
	components components
	dictionary map[string]*component
	logger     log.Logger
}

func (that *App) addComponent(component *component) {
	that.mx.Lock()
	defer that.mx.Unlock()

	that.components = append(that.components, component)
	that.dictionary[component.id] = component
}

func (that *App) Init(ctx context.Context) {
	for _, c := range that.components {
		if err := c.runInit(ctx, that); err != nil {
			panic(&componentError{c.name, fmt.Sprintf("init: %s", err.Error())})
		}
	}
}

func (that *App) Done(ctx context.Context) {
	ctx = context.WithValue(ctx, ApplicationContextKey, that)
	cs := that.components
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		c.runDone(ctx, that)
	}
}

func (that *App) Run(context.Context) {
	// nothing to do
}

func (that *App) get(ctx context.Context, name string, builder func(ctx context.Context, app *App) *component) *component {
	c := that.fetch(ctx, name)
	if c != nil {
		return c
	}

	c = builder(ctx, that)
	that.addComponent(c)

	return c
}

func (that *App) fetch(_ context.Context, name string) *component {
	that.mx.Lock()
	defer that.mx.Unlock()

	c, _ := that.dictionary[name]
	return c
}

type AppOptions struct {
	constructors []constructor
	logger       log.Logger
}

type AppOption func(opts *AppOptions)

func newConstructor[T any](constructor Constructor[T]) func(ctx context.Context) {
	return func(ctx context.Context) {
		_ = constructor(ctx)
	}
}

// WithAppService - add service to application
func WithAppService[T any](constructor ...Constructor[T]) AppOption {
	return func(opts *AppOptions) {
		for _, c := range constructor {
			opts.constructors = append(opts.constructors, newConstructor(c))
		}
	}
}

// WithAppDaemon - add daemon to application
func WithAppDaemon(daemons ...func(ctx context.Context)) AppOption {
	return func(opts *AppOptions) {
		for _, daemon := range daemons {
			opts.constructors = append(opts.constructors, daemon)
		}
	}
}

// WithAppLogger - set logger for application
func WithAppLogger(logger log.Logger) AppOption {
	return func(opts *AppOptions) {
		opts.logger = logger
	}
}

// Execute - primary entry point for build and run application
func Execute(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) {
	opts := buildAppOptions(options...)

	app, ctx := build(ctx, constructor, opts)
	app.Init(ctx)
	defer app.Done(ctx)

	app.Run(ctx)
}

// Build - build application
func Build(
	ctx context.Context,
	constructor Constructor[Application],
	options ...AppOption,
) (a Application, c context.Context) {
	opts := buildAppOptions(options...)
	return build(ctx, constructor, opts)
}

func build(
	ctx context.Context,
	constructor Constructor[Application],
	opts AppOptions,
) (a Application, c context.Context) {
	app := newApp(opts.logger)
	ctx = context.WithValue(ctx, ApplicationContextKey, app)

	application := constructor(ctx)

	for _, c := range opts.constructors {
		c(ctx)
	}

	return application, ctx
}

type ApplicationContextType int

var ApplicationContextKey ApplicationContextType = 0

// GetAppFromContext - get application from context
func GetAppFromContext(ctx context.Context) *App {
	app, _ := ctx.Value(ApplicationContextKey).(*App)
	if app == nil {
		panic(fmt.Errorf("application not found in context"))
	}
	return app
}

func buildAppOptions(options ...AppOption) AppOptions {
	opts := AppOptions{
		logger: log.NewDummyLogger(),
	}
	for _, o := range options {
		o(&opts)
	}
	return opts
}
