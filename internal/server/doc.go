// Package server wires together the HTTP server, routing, and middleware
// for the grpc-health-proxy sidecar.
//
// Usage:
//
//	cfg, _ := config.Parse()
//	srv, err := server.New(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Graceful shutdown on signal
//	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
//	defer stop()
//	go func() { _ = srv.Start() }()
//	<-ctx.Done()
//	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	_ = srv.Shutdown(shutCtx)
//
// Endpoints registered:
//
//	<HTTPPath>  — proxies the gRPC health check (configurable, default /healthz)
//	/livez      — always returns 200 OK (liveness probe)
package server
