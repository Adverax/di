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
	func() *MyApplication {
		return &MyApplication{
			App:        GetApp(),
			Events:     GetEvents(),
			Repository: GetRepository(),
		}
	},
)

var GetEvents = NewComponent(
	func() *MyEvents {
		return &MyEvents{}
	},
)

var GetRepository = NewComponent(
	func() *MyRepository {
		return &MyRepository{
			Events: GetEvents(),
		}
	},
)

var GetScheduler = NewComponent(
	func() *Scheduler {
		return &Scheduler{}
	},
	WithInit(func(instance *Scheduler) error {
		return instance.Start()
	}),
	AsWorker[*Scheduler](),
)

func TestDI(t *testing.T) {
	app, err := Build(GetApplication)
	require.NoError(t, err)
	defer app.Done()
	app.Start()
}
