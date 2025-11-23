package teamService

import "PullRequestService/domain"

type TeamService interface {
	PostTeam(team domain.Team) error
	GetTeam(teamName string) (domain.Team, error)
}

type teamService struct {
	repo TeamRepository
}

func NewTeamService(repo TeamRepository) TeamService {
	return &teamService{
		repo: repo,
	}
}

func (ts *teamService) PostTeam(team domain.Team) error {
	if err := ts.repo.PostTeam(team); err != nil {
		return err
	}
	return nil
}

func (ts *teamService) GetTeam(teamName string) (domain.Team, error) {
	team, err := ts.repo.GetTeam(teamName)
	if err != nil {
		return domain.Team{}, err
	}
	return team, nil
}
