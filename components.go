package di

import (
	"fmt"
	"reflect"
)

type Option[T any] func(options *Options[T])

type Options[T any] struct {
	init func(instance T) error
	done func()
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

type Builder[T any] interface {
	Build() (T, error)
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
	var active bool

	return func() T {
		if active {
			panic(fmt.Errorf("circular dependency detected"))
		}

		active = true
		defer func() {
			active = false
		}()

		if isZeroVal(instance) {
			instance = constructor()
			application.addComponent(newComponent(instance, opts))
		}

		return instance
	}
}

func NewComponentWithBuilder[T any](
	constructor func() Builder[T],
	options ...Option[T],
) func() T {
	var opts Options[T]
	for _, o := range options {
		o(&opts)
	}

	var instance T
	var active bool

	return func() T {
		if active {
			panic(fmt.Errorf("circular dependency detected"))
		}

		active = true
		defer func() {
			active = false
		}()

		if isZeroVal(instance) {
			builder := constructor()
			var err error
			instance, err = builder.Build()
			if err != nil {
				panic(err)
			}
			application.addComponent(newComponent(instance, opts))
		}

		return instance
	}
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
