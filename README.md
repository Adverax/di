# Adverax DI

[![Go Reference](https://pkg.go.dev/badge/github.com/adverax/di.svg)](https://pkg.go.dev/github.com/adverax/di)  
[![Go Report Card](https://goreportcard.com/badge/github.com/adverax/di)](https://goreportcard.com/report/github.com/adverax/di)  
[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)

Adverax DI is a lightweight and idiomatic dependency injection (DI) framework for Go, designed to be simple, efficient, and type-safe without relying on reflection or code generation.

## Key Features

- **Lightweight**: Minimalistic design, easy to integrate.
- **No Reflection**: Relies solely on Go's static typing, ensuring better performance and reliability.
- **No Code Generation**: Simplifies integration without requiring additional tools.
- **Type-Safe**: Eliminates the need for type casting, ensuring type safety at compile time.
- **Facilitates alive code**: Facilitates eliminating dead code in the IDE. Dead code is code that is never called. This is a common problem in large projects, where it is difficult to determine which code is used and which is not.
- **Helpful for graceful shutdown**: Facilitates graceful shutdown of the application. This is important for applications that need to release resources when they are no longer needed.
- **Support collections** of components for constructing multiple instances of the same type.

## Installation

Install the package using `go get`:

```bash
go get github.com/adverax/di
```

## Usage
```golang
package main

import (
	"context"
    "github.com/adverax/di"
)

type MyEvents struct {
	// Your events here
}

type MyRepository struct {
	// Dependency list
	Events *MyEvents
}

type MyScheduler struct {
}

func (s *MyScheduler) Start() error {
	// your code here
	return nil
}

type MyApplication struct {
	*di.App // Embedding App struct
	// Dependency list
	Events     *MyEvents
	Repository *MyRepository
}

// Declaration of components
var ComponentApplication = di.NewComponent(
	"MyApplication",
	func(ctx context.Context) (*MyApplication, error) {
		return &MyApplication{
			App:        di.GetAppFromContext(ctx),
			Events:     ComponentEvents(ctx),
			Repository: ComponentRepository(ctx),
		}, nil
	},
)

var ComponentEvents = di.NewComponent(
	"MyEvents",
	func(ctx context.Context) (*MyEvents, error) {
		return &MyEvents{}, nil
	},
)

var ComponentRepository = di.NewComponent(
	"MyRepository",
	func(ctx context.Context) (*MyRepository, error) {
		return &MyRepository{
			Events: ComponentEvents(ctx),
		}, nil
	},
)

var ComponentScheduler = di.NewComponent(
	"Scheduler",
	func(ctx context.Context) (*MyScheduler, error) {
		return &MyScheduler{}, nil
	},
	WithComponentInit(func(ctx context.Context, instance *Scheduler) error {
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
```

## Documentation
Detailed documentation is available on pkg.go.dev.

## Contributing
Contributions are welcome! Please open issues for bug reports or feature requests. Pull requests are encouraged.

## License
This project is licensed under the Apache 2.0 License. See the LICENSE file for details.
