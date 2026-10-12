package handlers

import (
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/services/depreport"
	"automation-hub/internal/services/telegram"
)

// RouteEnv holds what the webhook routes are built from.
type RouteEnv struct {
	Telegram         *telegram.Client
	DependencyReport *config.DependencyReportConfig
	Catalog          *messages.Catalog
	Location         *time.Location
	Sender           depreport.ReportSender
	Logger           *zap.Logger
}

// Route is a POST route: its path and the handler that serves it.
type Route struct {
	Path    string
	Handler http.Handler
}

// RouteFactory builds the route of a named webhook from its config entry.
type RouteFactory func(hook config.WebhookConfig, env RouteEnv) (Route, error)

// DependencyReportRoute is the name of the route that the dependency_report section configures.
const DependencyReportRoute = "dependency-report"

// routeFactories maps each webhook name the config may use to its factory.
var routeFactories = map[string]RouteFactory{
	"qbittorrent":         newTorrentRoute,
	DependencyReportRoute: newDependencyReportRoute,
}

// LookupRouteFactory returns the factory of a webhook name, and whether one exists for its source.
// fromSection is true for the dependency_report section and false for a hook entry. The
// dependency report comes only from its section, so a hook entry with that name finds no factory.
func LookupRouteFactory(name string, fromSection bool) (RouteFactory, bool) {
	if (name == DependencyReportRoute) != fromSection {
		return nil, false
	}
	factory, ok := routeFactories[name]
	return factory, ok
}

func newTorrentRoute(hook config.WebhookConfig, env RouteEnv) (Route, error) {
	handler := NewWebhookHandler(env.Telegram, hook.Config, env.Logger)
	return Route{Path: hook.Path, Handler: http.HandlerFunc(handler.HandleTorrentComplete)}, nil
}

// newDependencyReportRoute reads its path and settings from the dependency_report section, not from a hook entry.
func newDependencyReportRoute(_ config.WebhookConfig, env RouteEnv) (Route, error) {
	if env.DependencyReport == nil {
		return Route{}, errors.New("dependency_report section is not configured")
	}
	handler := NewDependencyReportHandler(*env.DependencyReport, env.Catalog, env.Location, env.Sender, env.Logger)
	return Route{Path: env.DependencyReport.Path, Handler: handler}, nil
}
