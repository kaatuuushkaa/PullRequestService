package statsService

import (
	"PullRequestService/domain"
	"PullRequestService/internal/web/stats"
	"gorm.io/gorm"
)

type StatsRepository interface {
	CountAssignmentsByUser() ([]stats.AssignmentsByUserItem, error)
	CountAssignmentsByPR() ([]stats.AssignmentsByPRItem, error)
	CountOpenPRs() (int, error)
	CountMergedPRs() (int, error)
}

type statsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) CountAssignmentsByUser() ([]stats.AssignmentsByUserItem, error) {
	var rows []stats.AssignmentsByUserItem

	err := r.db.
		Table("pull_request_reviewers").
		Select("reviewer_id as user_id, COUNT(*) as count").
		Group("reviewer_id").
		Scan(&rows).Error

	return rows, err
}

func (r *statsRepository) CountAssignmentsByPR() ([]stats.AssignmentsByPRItem, error) {
	var rows []stats.AssignmentsByPRItem

	err := r.db.
		Table("pull_request_reviewers").
		Select("pull_request_id, COUNT(*) as count").
		Group("pull_request_id").
		Scan(&rows).Error

	return rows, err
}

func (r *statsRepository) CountOpenPRs() (int, error) {
	var count int64
	err := r.db.Model(&domain.PullRequest{}).
		Where("status = ?", "OPEN").
		Count(&count).Error
	return int(count), err
}

func (r *statsRepository) CountMergedPRs() (int, error) {
	var count int64
	err := r.db.Model(&domain.PullRequest{}).
		Where("status = ?", "MERGED").
		Count(&count).Error
	return int(count), err
}
