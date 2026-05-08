// Package lifecycle provides structured startup, shutdown, and background-worker
// management for the portwatch daemon.
//
// # Manager
//
// Manager accepts OnStart and OnStop hooks. Call Run to execute all start hooks
// in registration order, block until a SIGINT/SIGTERM or context cancellation,
// then execute stop hooks in reverse (LIFO) order within the configured timeout.
//
//	manager := lifecycle.New(5 * time.Second)
//	manager.OnStart("scanner", scanner.Start)
//	manager.OnStop("scanner", scanner.Stop)
//	if err := manager.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// # Runner
//
// Runner launches a set of Worker implementations concurrently. If any worker
// returns a non-nil error the shared context is cancelled so remaining workers
// can exit gracefully.
//
//	runner := lifecycle.NewRunner()
//	runner.Add(monitorWorker)
//	runner.Add(metricsWorker)
//	if err := runner.Run(ctx); err != nil {
//	    log.Printf("runner error: %v", err)
//	}
package lifecycle
