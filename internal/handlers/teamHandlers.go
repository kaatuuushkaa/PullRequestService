package handlers

import (
	"PullRequestService/domain"
	"PullRequestService/internal/teamService"
	"PullRequestService/internal/web/teams"
	"context"
	"fmt"
)

type TeamHandler struct {
	service teamService.TeamService
}

func NewTeamHandler(service teamService.TeamService) *TeamHandler {
	return &TeamHandler{
		service: service,
	}
}

func (t *TeamHandler) GetTeamGet(ctx context.Context, request teams.GetTeamGetRequestObject) (teams.GetTeamGetResponseObject, error) {

	team, err := t.service.GetTeam(request.Params.TeamName)
	if err != nil {
		return teams.GetTeamGet404JSONResponse{
			Error: struct {
				Code    teams.ErrorResponseErrorCode `json:"code"`
				Message string                       `json:"message"`
			}{
				Code:    teams.NOTFOUND,
				Message: "resource not found",
			},
		}, nil
	}

	members := make([]teams.TeamMember, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, teams.TeamMember{
			UserId:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	response := teams.GetTeamGet200JSONResponse{
		TeamName: team.Name,
		Members:  members,
	}

	return response, nil

}

func (t *TeamHandler) PostTeamAdd(ctx context.Context, request teams.PostTeamAddRequestObject) (teams.PostTeamAddResponseObject, error) {
	req := request.Body

	members := make([]domain.User, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, domain.User{
			ID:       m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team := domain.Team{
		Name:    req.TeamName,
		Members: members,
	}

	err := t.service.PostTeam(team)
	if err != nil {
		return teams.PostTeamAdd400JSONResponse{
			Error: struct {
				Code    teams.ErrorResponseErrorCode `json:"code"`
				Message string                       `json:"message"`
			}{
				Code:    teams.TEAMEXISTS,
				Message: fmt.Sprintf("%s already exists", team.Name),
			},
		}, nil
	}

	response := teams.PostTeamAdd201JSONResponse{
		Team: &teams.Team{
			TeamName: team.Name,
			Members:  req.Members,
		},
	}
	return response, nil
}
