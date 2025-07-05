package bridgr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	serverReadHeaderTimeout = 5 * time.Second
)

var (
	logFmt          = "2/Jan/2006:15:04:05 -0700"
	shutdownTimeout = 10 * time.Second
	done            = make(chan bool)
	quit            = make(chan os.Signal, 1)
)

type loggingResponseWriter struct {
	status int
	body   string
	http.ResponseWriter
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	w.body = string(body)
	return w.ResponseWriter.Write(body)
}

func shutdownHandler(server *http.Server) {
	<-quit
	fmt.Print("Server is shutting down... ")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}

// Serve starts and runs the static web server for Bridgr
func Serve(addr string, root http.FileSystem) error {
	if _, err := root.Open("/"); err != nil {
		return fmt.Errorf("filesystem %s does not exist", root)
	}
	fs := http.FileServer(root)
	handlerChain := logMiddleware(customHeaders(fs))
	http.Handle("/", handlerChain)
	server := &http.Server{Addr: addr, ReadHeaderTimeout: serverReadHeaderTimeout, Handler: http.DefaultServeMux}

	signal.Notify(quit, os.Interrupt)
	go shutdownHandler(server)

	fmt.Printf("Starting Bridgr HTTP server on %s for static directory %s\n", addr, root)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", addr, err)
	}

	<-done
	fmt.Println("stopped")
	return nil
}

func customHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Bridgr/"+Version)
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now().Format(logFmt)
		loggingRW := &loggingResponseWriter{
			ResponseWriter: w,
		}
		next.ServeHTTP(loggingRW, r) // continue with the response

		// clog (used by the rest of bridgr) doesn't support changing the "prefix" of logging output
		// so, we use direct Printf here instead so the logs aren't "double formatted" - first by clog and then in CLF form
		if Verbose {
			fmt.Printf("%s - - [%s] %q %d %d %q %s\n", r.RemoteAddr, timestamp, r.Method+" "+r.URL.Path+" "+r.Proto, loggingRW.status, len(loggingRW.body), r.Referer(), r.UserAgent())
		}
	})
}
