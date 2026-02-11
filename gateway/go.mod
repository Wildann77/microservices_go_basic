module github.com/microservices-go/gateway

go 1.24.0

require (
	github.com/99designs/gqlgen v0.17.86
	github.com/go-chi/chi/v5 v5.0.10
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/microservices-go/shared v0.0.0
	github.com/vektah/gqlparser/v2 v2.5.31
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.31.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)

replace github.com/microservices-go/shared => ../shared
