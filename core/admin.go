// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

import (
	"bytes"
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/goburrow/gol"
	"github.com/goburrow/health"
)

const (
	metricsUri     = "/metrics"
	pingUri        = "/ping"
	runtimeUri     = "/runtime"
	healthCheckUri = "/healthcheck"
	tasksUri       = "/tasks"

	adminHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Operational Menu</title>
</head>
<body>
	<h1>Operational Menu</h1>
	<ul>%[1]s</ul>
</body>
</html>
`
	noHealthChecksWarning = `
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!    THIS APPLICATION HAS NO HEALTHCHECKS.    !
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
`

	adminLoggerName = "gomelon.admin"

	gcTaskName  = "gc"
	logTaskName = "log"
)

type AdminHandler interface {
	Path() string
	Name() string
	http.Handler
}

type AdminEnvironment struct {
	ServerHandler ServerHandler
	HealthChecks  health.Registry

	handlers []AdminHandler
	tasks    []Task
}

func NewAdminEnvironment() *AdminEnvironment {
	env := &AdminEnvironment{
		HealthChecks: health.NewRegistry(),
	}
	// Default handlers
	env.AddHandler(&metricsHandler{}, &pingHandler{}, &runtimeHandler{}, &healthCheckHandler{env.HealthChecks})
	// Default tasks
	env.AddTask(&gcTask{}, &logTask{})
	return env
}

// AddTask adds a new task to admin environment. AddTask is not concurrent-safe.
func (env *AdminEnvironment) AddTask(task ...Task) {
	env.tasks = append(env.tasks, task...)
}

// AddHandler registers a handler entry for admin page.
func (env *AdminEnvironment) AddHandler(handler ...AdminHandler) {
	env.handlers = append(env.handlers, handler...)
}

// onStarting registers all required HTTP handlers
func (env *AdminEnvironment) onStarting() {
	env.ServerHandler.Handle("GET", "/", &adminHomeHandler{
		handlers:    env.handlers,
		contextPath: env.ServerHandler.PathPrefix(),
	})
	for _, h := range env.handlers {
		env.ServerHandler.Handle("GET", h.Path(), h)
	}

	for _, task := range env.tasks {
		path := tasksUri + "/" + task.Name()
		env.ServerHandler.Handle("POST", path, task)
	}
	env.logTasks()
	env.logHealthChecks()
}

func (env *AdminEnvironment) onStopped() {
}

// logTasks prints all registered tasks to the log
func (env *AdminEnvironment) logTasks() {
	var buf bytes.Buffer
	for _, task := range env.tasks {
		fmt.Fprintf(&buf, "    %-7s %s%s/%s (%T)\n", "POST",
			env.ServerHandler.PathPrefix(), tasksUri, task.Name(), task)
	}
	gol.GetLogger(adminLoggerName).Info("tasks =\n\n%s", buf.String())
}

// logTasks prints all registered tasks to the log
func (env *AdminEnvironment) logHealthChecks() {
	logger := gol.GetLogger(adminLoggerName)
	names := env.HealthChecks.Names()
	if len(names) <= 0 {
		logger.Warn(noHealthChecksWarning)
	}
	logger.Debug("health checks = %v", names)
}

// AdminHandler implement http.Handler
type adminHomeHandler struct {
	handlers    []AdminHandler
	contextPath string
}

// ServeHTTP handles request to the root of Admin page
func (handler *adminHomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	for _, h := range handler.handlers {
		fmt.Fprintf(&buf, "<li><a href=\"%[1]s%[2]s\">%[3]s</a></li>",
			handler.contextPath, h.Path(), h.Name())
	}

	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintf(w, adminHTML, buf.String())
}

// healthCheckHandler is the http handler for /healthcheck page
type healthCheckHandler struct {
	registry health.Registry
}

func (handler *healthCheckHandler) Name() string {
	return "Healthcheck"
}

func (handler *healthCheckHandler) Path() string {
	return healthCheckUri
}

func (handler *healthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	results := handler.registry.RunHealthChecks()

	if len(results) == 0 {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("No health checks registered."))
		return
	}
	if !isAllHealthy(results) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for name, result := range results {
		fmt.Fprintf(w, "%s:\n\tHealthy: %t\n", name, result.Healthy)
		if result.Message != "" {
			fmt.Fprintf(w, "\tMessage: %s\n", result.Message)
		}
		if result.Cause != nil {
			fmt.Fprintf(w, "\tCause: %+v\n", result.Cause)
		}
	}
}

// isAllHealthy checks if all are healthy
func isAllHealthy(results map[string]*health.Result) bool {
	for _, result := range results {
		if !result.Healthy {
			return false
		}
	}
	return true
}

// metricsHandler displays expvars.
type metricsHandler struct {
}

func (handler *metricsHandler) Name() string {
	return "Metrics"
}

func (handler *metricsHandler) Path() string {
	return metricsUri
}

func (*metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "}")
}

// pingHandler handles ping request to admin /ping
type pingHandler struct {
}

func (handler *pingHandler) Name() string {
	return "Ping"
}

func (handler *pingHandler) Path() string {
	return pingUri
}

func (handler *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
}

// runtimeHandler displays runtime statistics.
type runtimeHandler struct {
}

func (handler *runtimeHandler) Name() string {
	return "Runtime"
}

func (handler *runtimeHandler) Path() string {
	return runtimeUri
}

func (handler *runtimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "NumCPU: %d\nNumCgoCall: %d\nNumGoroutine: %d\n",
		runtime.NumCPU(), runtime.NumCgoCall(), runtime.NumGoroutine())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// General statistics
	fmt.Fprintf(w, "MemStats:\n\tAlloc: %d\n\tTotalAlloc: %d\n\tSys: %d\n\tLookups: %d\n\tMallocs: %d\n\tFrees: %d\n",
		m.Alloc, m.TotalAlloc, m.Sys, m.Lookups, m.Mallocs, m.Frees)
	// Main allocation heap statistics
	fmt.Fprintf(w, "\tHeapAlloc: %d\n\tHeapSys: %d\n\tHeapIdle: %d\n\tHeapInuse: %d\n\tHeapReleased: %d\n\tHeapObjects: %d\n",
		m.HeapAlloc, m.HeapSys, m.HeapIdle, m.HeapInuse, m.HeapReleased, m.HeapObjects)
	// Low-level fixed-size structure allocator statistics
	fmt.Fprintf(w, "\tStackInuse: %d\n\tStackSys: %d\n\tMSpanInuse: %d\n\tMSpanSys: %d\n\tMCacheInuse: %d\n\tMCacheSys: %d\n\tBuckHashSys: %d\n\tGCSys: %d\n\tOtherSys: %d\n",
		m.StackInuse, m.StackSys, m.MSpanInuse, m.MSpanSys, m.MCacheInuse, m.MCacheSys, m.BuckHashSys, m.GCSys, m.OtherSys)
	// Garbage collector statistics
	fmt.Fprintf(w, "\tNextGC: %d\n\tLastGC: %d\n\tPauseTotalNs: %d\n\tNumGC: %d\n\tEnableGC: %t\n\tDebugGC: %t\n",
		m.NextGC, m.LastGC, m.PauseTotalNs, m.NumGC, m.EnableGC, m.DebugGC)

	fmt.Fprintf(w, "Version: %s\n", runtime.Version())
}

// gcTask performs a garbage collection
type gcTask struct {
}

func (*gcTask) Name() string {
	return gcTaskName
}

func (*gcTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Running GC...\n"))
	runtime.GC()
	w.Write([]byte("Done!\n"))
}

// logTask gets and sets logger level
type logTask struct {
}

func (*logTask) Name() string {
	return logTaskName
}

func (*logTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// Can have multiple loggers
	loggers, ok := query["logger"]
	if !ok || len(loggers) == 0 {
		return
	}
	// But only one level
	level := query.Get("level")
	if level != "" {
		logLevel, ok := parseLogLevel(level)
		if !ok {
			http.Error(w, "Level is not supported", http.StatusBadRequest)
			return
		}
		for _, name := range loggers {
			logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
			if ok {
				logger.SetLevel(logLevel)
			}
		}
	}
	// Print level of each logger
	for _, name := range loggers {
		logger, ok := gol.GetLogger(name).(*gol.DefaultLogger)
		if ok {
			fmt.Fprintf(w, "%s: %s\n", name, gol.LevelString(logger.Level()))
		}
	}
}

// parseLogLevel returns respective gol.Level of the given string
func parseLogLevel(level string) (gol.Level, bool) {
	// Changing log level is not executed regularly so it's not worth having
	// logLevels in static scope
	var logLevels = []gol.Level{
		gol.LevelAll,
		gol.LevelTrace,
		gol.LevelDebug,
		gol.LevelInfo,
		gol.LevelWarn,
		gol.LevelError,
		gol.LevelOff,
	}
	for _, l := range logLevels {
		if strings.EqualFold(level, gol.LevelString(l)) {
			return l, true
		}
	}
	return gol.LevelOff, false
}