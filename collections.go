package di

import (
	"context"
	"fmt"
)

type Collection[T any] []T
type CollectionItemBuilder[W any, T any] func(ctx context.Context, instance T) (W, error)
type CollectionConstructor[T any] func(ctx context.Context) Collection[T]

type CollectionOptions[W any, T any] struct {
	items   []Constructor[T]
	options []Option[W]
}

type CollectionOption[W any, T any] func(options *CollectionOptions[W, T])

func WithCollectionItemOptions[W any, T any](opts ...Option[W]) CollectionOption[W, T] {
	return func(options *CollectionOptions[W, T]) {
		options.options = append(options.options, opts...)
	}
}

func WithCollectionItems[W any, T any](items ...Constructor[T]) CollectionOption[W, T] {
	return func(options *CollectionOptions[W, T]) {
		options.items = append(options.items, items...)
	}
}

// NewCollection makes new collection of items by wrapping them by builder func
func NewCollection[W any, T any](
	name string,
	builder CollectionItemBuilder[W, T],
	options ...CollectionOption[W, T],
) CollectionConstructor[W] {
	var opts CollectionOptions[W, T]
	for _, opt := range options {
		opt(&opts)
	}

	b := func(ctx context.Context) (Collection[W], error) {
		var list Collection[W]
		for i, item := range opts.items {
			itm := NewComponent(
				fmt.Sprintf("%s:%d", name, i+1),
				func(ctx context.Context) (W, error) {
					w := item(ctx)
					return builder(ctx, w)
				},
				opts.options...,
			)
			list = append(list, itm(ctx))
		}
		return list, nil
	}

	c := controller[Collection[W]]{
		options: Options[Collection[W]]{
			name: name,
		},
		builder: b,
		id:      enumerator(),
	}

	return c.get
}
