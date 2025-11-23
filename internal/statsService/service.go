package statsService

import (
	"PullRequestService/internal/web/stats"
)

type StatsService interface {
	GetStats() (stats.StatsResponse, error)
}

type statsService struct {
	repo StatsRepository
}

func NewStatsService(repo StatsRepository) StatsService {
	return &statsService{repo: repo}
}

func (s *statsService) GetStats() (stats.StatsResponse, error) {
	byUser, err := s.repo.CountAssignmentsByUser()
	if err != nil {
		return stats.StatsResponse{}, err
	}

	byPR, err := s.repo.CountAssignmentsByPR()
	if err != nil {
		return stats.StatsResponse{}, err
	}

	openCount, err := s.repo.CountOpenPRs()
	if err != nil {
		return stats.StatsResponse{}, err
	}

	mergedCount, err := s.repo.CountMergedPRs()
	if err != nil {
		return stats.StatsResponse{}, err
	}

	return stats.StatsResponse{
		AssignmentsByUser: byUser,
		AssignmentsByPr:   byPR,
		OpenPrsCount:      openCount,
		MergedPrsCount:    mergedCount,
	}, nil
}
