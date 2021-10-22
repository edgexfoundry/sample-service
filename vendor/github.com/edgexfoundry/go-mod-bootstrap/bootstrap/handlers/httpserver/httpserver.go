/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package httpserver

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/gorilla/mux"
)

// HttpServer contains references to dependencies required by the http server implementation.
type HttpServer struct {
	router           *mux.Router
	isRunning        bool
	doListenAndServe bool
}

// NewBootstrap is a factory method that returns an initialized HttpServer receiver struct.
func NewBootstrap(router *mux.Router, doListenAndServe bool) *HttpServer {
	return &HttpServer{
		router:           router,
		isRunning:        false,
		doListenAndServe: doListenAndServe,
	}
}

// IsRunning returns whether or not the http server is running.  It is provided to support delayed shutdown of
// any resources required to successfully process http requests until after all outstanding requests have been
// processed (e.g. a database connection).
func (b *HttpServer) IsRunning() bool {
	return b.isRunning
}

// BootstrapHandler fulfills the BootstrapHandler contract.  It creates two go routines -- one that executes ListenAndServe()
// and another that waits on closure of a context's done channel before calling Shutdown() to cleanly shut down the
// http server.
func (b *HttpServer) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {

	lc := container.LoggingClientFrom(dic.Get)

	if !b.doListenAndServe {
		lc.Info("Web server intentionally NOT started.")
		wg.Add(1)
		go func() {
			defer wg.Done()

			b.isRunning = true
			<-ctx.Done()
			b.isRunning = false
		}()
		return true

	}

	bootstrapConfig := container.ConfigurationFrom(dic.Get).GetBootstrap()

	// this allows env override to explicitly set the value used
	// for ListenAndServe as needed for different deployments
	port := strconv.Itoa(bootstrapConfig.Service.Port)
	addr := bootstrapConfig.Service.ServerBindAddr + ":" + port
	// for backwards compatibility, the Host value is the default value if
	// the ServerBindAddr value is not specified
	if bootstrapConfig.Service.ServerBindAddr == "" {
		addr = bootstrapConfig.Service.Host + ":" + port
	}

	timeout := time.Millisecond * time.Duration(bootstrapConfig.Service.Timeout)
	server := &http.Server{
		Addr:         addr,
		Handler:      b.router,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	lc.Info("Web server starting (" + addr + ")")

	wg.Add(1)
	go func() {
		defer wg.Done()

		b.isRunning = true
		_ = server.ListenAndServe()
		lc.Info("Web server stopped")
		b.isRunning = false
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		lc.Info("Web server shutting down")
		_ = server.Shutdown(context.Background())
		lc.Info("Web server shut down")
	}()

	return true
}
