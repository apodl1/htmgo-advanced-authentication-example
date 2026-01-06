package main

import (
	"advancedauth/__htmgo"
	"advancedauth/internal/db"
	"context"
	"fmt"
	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"
	"io/fs"
	"net/http"
	"time"
)

func main() {
	locator := service.NewLocator()

	// Initialize queries once to share between the app and the background worker
	queries := db.Provide()

	service.Set(locator, service.Singleton, func() *db.Queries {
		return queries
	})

	// Start the background cleanup job
	go startSessionCleanup(queries)

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			http.FileServerFS(sub)

			app.Router.Handle("/public/*", http.StripPrefix("/public", http.FileServerFS(sub)))
			__htmgo.Register(app.Router)
		},
	})
}

func startSessionCleanup(queries *db.Queries) {
	// Run once a day. Adjust the duration if needed
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	fmt.Println("Session cleanup worker started")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

		fmt.Printf("[%s] Running expired session cleanup...\n", time.Now().Format(time.RFC3339))

		err := queries.DeleteExpiredSessions(ctx)
		if err != nil {
			fmt.Printf("Error deleting expired sessions: %v\n", err)
		}

		err = queries.DeleteExpiredRememberTokens(ctx)
		if err != nil {
			fmt.Printf("Error deleting expired remember tokens: %v\n", err)
		}

		cancel()
	}
}
