package app

import (
	"github.com/l1f/mahlzeit/db/queries"
	"github.com/l1f/mahlzeit/internal/templates"
	"go.uber.org/zap"
)

// Application holds the required resources for the application.
// This includes, but is not limited, to loggers, databases, cache handlers, etc.
type Application struct {
	Templates templates.Templates
	Queries   *queries.Queries
	Logger    *zap.Logger
}

type Configuration struct {
	Database struct {
		ConnectionString string `toml:"connection-string"`
	} `toml:"database"`
	Web struct {
		Endpoint    string `toml:"endpoint"`
		TemplateDir string `toml:"template-dir"`
	}
}
