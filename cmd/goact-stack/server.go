package main

import (
	"io/fs"
	"net/http"
	"strings"

	"goact-stack/internal/static"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runServer(cmd *cobra.Command, args []string) error {
	port := viper.GetString("port")
	logLevel := viper.GetString("log_level")

	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	if logLevel == "debug" {
		e.Use(middleware.Logger())
	} else {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
		}))
	}

	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())

	// API routes
	api := e.Group("/api")
	api.GET("/health", healthHandler)

	// Serve embedded frontend assets
	distFS, err := fs.Sub(static.DistFS, "dist")
	if err != nil {
		return err
	}

	// Static file handler with SPA fallback
	fileServer := http.FileServer(http.FS(distFS))
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the file directly
		if path != "/" && !strings.HasSuffix(path, "/") {
			// Check if file exists in embedded FS
			if f, err := distFS.Open(strings.TrimPrefix(path, "/")); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})))

	e.Logger.Printf("Starting server on %s", port)
	return e.Start(port)
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
