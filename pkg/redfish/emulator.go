package redfish

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtbmc/pkg/generated/redfish/server"
	"kubevirt.io/kubevirtbmc/pkg/resourcemanager"
)

const (
	defaultUserName = "admin"
	defaultPassword = "password"
)

type Emulator struct {
	ctx    context.Context
	port   int
	wg     sync.WaitGroup
	server *http.Server
}

func NewEmulator(ctx context.Context, port int, resourceManager resourcemanager.ResourceManager) *Emulator {
	apiService := NewAPIService(resourceManager)
	apiController := server.NewDefaultAPIController(apiService)
	router := server.NewRouter(BasicAuthMiddleware, apiController)

	return &Emulator{
		ctx:  ctx,
		port: port,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
	}
}

func (e *Emulator) Run() error {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()

	return nil
}

func (e *Emulator) Stop() {
	if err := e.server.Shutdown(e.ctx); err != nil {
		fmt.Println(err)
	}
	e.wg.Wait()
	logrus.Info("Redfish emulator gracefully stopped")
}

func BasicAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO add support for sessions also?
		username, password, exists := r.BasicAuth()
		if !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if username != defaultUserName || password != defaultPassword {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
