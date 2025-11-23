package app

import (
	"PullRequestService/internal/db"
	"PullRequestService/internal/handlers"
	"PullRequestService/internal/pullRequestService"
	"PullRequestService/internal/statsService"
	"PullRequestService/internal/teamService"
	"PullRequestService/internal/userService"
	"PullRequestService/internal/web/pullRequests"
	"PullRequestService/internal/web/stats"
	"PullRequestService/internal/web/teams"
	"PullRequestService/internal/web/users"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	E *echo.Echo
}

func New() (*App, error) {
	dbConn, err := db.InitDB()
	if err != nil {
		return nil, err
	}

	e := echo.New()

	e.Use(middleware.Logger())

	repoTeam := teamService.NewTeamRepository(dbConn)
	serviceTeam := teamService.NewTeamService(repoTeam)
	handlerTeam := handlers.NewTeamHandler(serviceTeam)
	teams.RegisterHandlers(e, teams.NewStrictHandler(handlerTeam, nil))

	repoUser := userService.NewUserRepository(dbConn)
	serviceUser := userService.NewUserService(repoUser)
	handlerUser := handlers.NewUserHandler(serviceUser)
	users.RegisterHandlers(e, users.NewStrictHandler(handlerUser, nil))

	repoPR := pullRequestService.NewPullRequestRepository(dbConn)
	servicePR := pullRequestService.NewPullRequestService(repoPR)
	handlerPR := handlers.NewPullRequestHandler(servicePR)
	pullRequests.RegisterHandlers(e, pullRequests.NewStrictHandler(handlerPR, nil))

	repoStat := statsService.NewStatsRepository(dbConn)
	serviceStat := statsService.NewStatsService(repoStat)
	handlerStat := handlers.NewStatsHandler(serviceStat)
	stats.RegisterHandlers(e, stats.NewStrictHandler(handlerStat, nil))

	return &App{E: e}, nil
}
