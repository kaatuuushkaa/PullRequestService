package main

import (
	"PullRequestService/internal/db"
	"PullRequestService/internal/handlers"
	"PullRequestService/internal/pullRequestService"
	"PullRequestService/internal/teamService"
	"PullRequestService/internal/userService"
	"PullRequestService/internal/web/pullRequests"
	"PullRequestService/internal/web/teams"
	"PullRequestService/internal/web/users"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	repoTeam := teamService.NewTeamRepository(database)
	serviceTeam := teamService.NewTeamService(repoTeam)
	handlerTeam := handlers.NewTeamHandler(serviceTeam)

	repoUser := userService.NewUserRepository(database)
	serviceUser := userService.NewUserService(repoUser)
	handlerUser := handlers.NewUserHandler(serviceUser)

	repoPR := pullRequestService.NewPullRequestRepository(database)
	servicePR := pullRequestService.NewPullRequestService(repoPR)
	handlerPR := handlers.NewPullRequestHandler(servicePR)

	e := echo.New()

	e.Use(middleware.Logger())

	strictHandlerTeams := teams.NewStrictHandler(handlerTeam, nil)
	teams.RegisterHandlers(e, strictHandlerTeams)

	strictHandlerPullRequests := pullRequests.NewStrictHandler(handlerPR, nil)
	pullRequests.RegisterHandlers(e, strictHandlerPullRequests)

	strictHandlerUser := users.NewStrictHandler(handlerUser, nil)
	users.RegisterHandlers(e, strictHandlerUser)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("failed to start with err: %v", err)
	}

}
