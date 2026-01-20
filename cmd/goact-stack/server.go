package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"goact-stack/internal/static"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runServer(cmd *cobra.Command, args []string) error {
	port := viper.GetString("port")

	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
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
	e.GET("/*", func(c *echo.Context) error {
		path := c.Request().URL.Path

		// Try to serve the file directly
		if path != "/" && !strings.HasSuffix(path, "/") {
			// Check if file exists in embedded FS
			if f, err := distFS.Open(strings.TrimPrefix(path, "/")); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Response(), c.Request())
				return nil
			}
		}

		// Fallback to index.html for SPA routing
		c.Request().URL.Path = "/"
		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	slog.Info("Starting server", "port", port)
	return e.Start(port)
}

func healthHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
