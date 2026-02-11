package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/microservices-go/gateway/graph"
	"github.com/microservices-go/gateway/graph/generated"
	"github.com/microservices-go/gateway/middleware"
	"github.com/microservices-go/shared/config"
	sharedMiddleware "github.com/microservices-go/shared/middleware"
)

const defaultPort = "4000"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Get service URLs
	userServiceURL, orderServiceURL, paymentServiceURL := graph.GetServiceURLs()

	// Load JWT config
	jwtConfig := config.LoadJWTConfig()

	// Create resolver
	resolver := graph.NewResolver(userServiceURL, orderServiceURL, paymentServiceURL)

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Add transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Add extensions
	srv.Use(extension.Introspection{})
	srv.Use(extension.FixedComplexityLimit(100))

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(sharedMiddleware.LoggingMiddleware)
	r.Use(sharedMiddleware.RecoveryMiddleware)
	r.Use(sharedMiddleware.SecurityHeadersMiddleware)
	r.Use(sharedMiddleware.CORSMiddleware([]string{"*"}))

	// DataLoader middleware
	r.Use(graph.DataLoaderMiddleware(userServiceURL, orderServiceURL, paymentServiceURL))

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtConfig)
	r.Use(authMiddleware.Middleware)

	// Routes
	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", srv)

	log.Printf("Connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}