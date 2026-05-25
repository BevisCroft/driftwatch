// Package graceful provides a Manager that coordinates orderly shutdown of
// driftwatch daemon components in response to OS signals (SIGINT, SIGTERM).
//
// Components register a ShutdownFunc via Register. When a termination signal
// is received (or Shutdown is called directly), each handler is invoked in
// reverse registration order (LIFO) so that higher-level components shut down
// before the lower-level resources they depend on.
//
// Usage:
//
//	gm := graceful.New(10*time.Second, nil)
//	gm.Register(func(ctx context.Context) error {
//		return server.Shutdown(ctx)
//	})
//	if err := gm.Wait(); err != nil {
//		log.Printf("shutdown error: %v", err)
//	}
package graceful
