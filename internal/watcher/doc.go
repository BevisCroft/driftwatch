// Package watcher implements a periodic polling loop that loads service
// manifests from a directory, runs drift detection on each manifest, and
// forwards the results to a reporter.
//
// Basic usage:
//
//	r := reporter.New(os.Stdout, "text")
//	w := watcher.New("/etc/driftwatch/manifests", 30*time.Second, r)
//	if err := w.Run(ctx); err != nil && err != context.Canceled {
//		log.Fatal(err)
//	}
//
// The watcher respects context cancellation and returns ctx.Err() when
// the context is done, making it straightforward to integrate with
// signal-based shutdown logic.
package watcher
