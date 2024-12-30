package di

import (
	"context"
	"testing"
)

type MyEvents struct {
	// Your events here
}

type MyRepository struct {
	// Dependency list
	Events *MyEvents
}

type Scheduler struct {
}

func (s *Scheduler) Start() error {
	// your code here
	return nil
}

type MyApplication struct {
	*App // Embedding App struct
	// Dependency list
	Events     *MyEvents
	Repository *MyRepository
}

// Declaration of components
var GetApplication = NewComponent(
	"MyApplication",
	func(ctx context.Context) (Application, error) {
		return &MyApplication{
			App:        GetAppFromContext(ctx),
			Events:     GetEvents(ctx),
			Repository: GetRepository(ctx),
		}, nil
	},
)

var GetEvents = NewComponent(
	"MyEvents",
	func(ctx context.Context) (*MyEvents, error) {
		return &MyEvents{}, nil
	},
)

var GetRepository = NewComponent(
	"MyRepository",
	func(ctx context.Context) (*MyRepository, error) {
		return &MyRepository{
			Events: GetEvents(ctx),
		}, nil
	},
)

var GetScheduler = NewComponent(
	"Scheduler",
	func(ctx context.Context) (*Scheduler, error) {
		return &Scheduler{}, nil
	},
	WithComponentInit(func(ctx context.Context, instance *Scheduler) error {
		return instance.Start()
	}),
)

func TestDI(t *testing.T) {
	Execute(
		context.Background(),
		GetApplication,
		WithAppService(GetScheduler),
	)
}
