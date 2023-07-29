package di

import (
	"github.com/stretchr/testify/require"
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
	App // Embedding App struct
	// Dependency list
	Events     *MyEvents
	Repository *MyRepository
}

func (a *MyApplication) Start() {
	// your code here
}

// Declaration of components
var GetApplication = NewComponent(
	func() (*MyApplication, error) {
		return &MyApplication{
			App:        GetApp(),
			Events:     GetEvents(),
			Repository: GetRepository(),
		}, nil
	},
)

var GetEvents = NewComponent(
	func() (*MyEvents, error) {
		return &MyEvents{}, nil
	},
)

var GetRepository = NewComponent(
	func() (*MyRepository, error) {
		return &MyRepository{
			Events: GetEvents(),
		}, nil
	},
)

var GetScheduler = NewComponent(
	func() (*Scheduler, error) {
		return &Scheduler{}, nil
	},
	WithInit(func(instance *Scheduler) error {
		return instance.Start()
	}),
)

func TestDI(t *testing.T) {
	app, err := Build(
		GetApplication,
		WithWorker(GetScheduler),
	)
	require.NoError(t, err)
	defer app.Done()
	app.Start()
}
