package di

import (
	"context"
	"fmt"
)

// Define components
type MyEvents struct {
	// Your events here
}

type MyRepository struct {
	Events *MyEvents
}

type MyScheduler struct{}

func (s *MyScheduler) Start() error {
	fmt.Println("Scheduler started")
	return nil
}

type MyApplication struct {
	*App
	Events     *MyEvents
	Repository *MyRepository
}

// Declare components
var ComponentApplication = NewComponent(
	"MyApplication",
	func(ctx context.Context) (Application, error) {
		return &MyApplication{
			App:        GetAppFromContext(ctx),
			Events:     ComponentEvents(ctx),
			Repository: ComponentRepository(ctx),
		}, nil
	},
)

var ComponentEvents = NewComponent(
	"MyEvents",
	func(ctx context.Context) (*MyEvents, error) {
		return &MyEvents{}, nil
	},
)

var ComponentRepository = NewComponent(
	"MyRepository",
	func(ctx context.Context) (*MyRepository, error) {
		return &MyRepository{
			Events: ComponentEvents(ctx),
		}, nil
	},
)

var ComponentScheduler = NewComponent(
	"Scheduler",
	func(ctx context.Context) (*MyScheduler, error) {
		return &MyScheduler{}, nil
	},
	WithComponentInit(func(ctx context.Context, instance *MyScheduler) error {
		return instance.Start()
	}),
)

// Example of dependency injection usage
func Example() {
	// Execute application
	Execute(
		context.Background(),
		ComponentApplication,
		WithAppService(ComponentScheduler),
	)

	// Output:
	// Scheduler started
}
