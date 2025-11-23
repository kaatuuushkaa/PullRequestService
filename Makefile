DB_DSN := postgres://postgres:yourpassword@localhost:5432/postgres?sslmode=disable
MIGRATE := migrate -path ./migrations -database $(DB_DSN)

run:
	go run cmd/main.go

migrate-new-teams:
	migrate create -ext sql -dir ./migrations teams

migrate-new-users:
	migrate create -ext sql -dir ./migrations users

migrate-new-pullRequests:
	migrate create -ext sql -dir ./migrations pullRequests

migrate:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

gen-teams:
	oapi-codegen -config openapi/.openapi -include-tags Teams -package teams openapi/openapi.yaml > ./internal/web/teams/api.gen.go

gen-users:
	oapi-codegen -config openapi/.openapi -include-tags Users -package users openapi/openapi.yaml > ./internal/web/users/api.gen.go

gen-pullRequests:
	oapi-codegen -config openapi/.openapi -include-tags PullRequests -package pullRequests openapi/openapi.yaml > ./internal/web/pullRequests/api.gen.go

gen-stats:
	oapi-codegen -config openapi/.openapi -include-tags Stats -package stats openapi/openapi.yaml > ./internal/web/stats/api.gen.go

gen: gen-teams gen-users gen-pullRequests gen-stats

lint:
	golangci-lint run --out-format=colored-line-number