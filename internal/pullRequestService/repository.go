package pullRequestService

import (
	"PullRequestService/domain"
	"gorm.io/gorm"
)

type PullRequestRepository interface {
	CreatePR(pr domain.PullRequest) error
	GetPR(id string) (domain.PullRequest, error)
	GetTeamMembers(teamName string) ([]domain.User, error)
	GetUserByID(userID string) (domain.User, error)
	UpdatePR(pr domain.PullRequest) error
}

type pullRequestRepository struct {
	db *gorm.DB
}

func NewPullRequestRepository(db *gorm.DB) PullRequestRepository {
	return &pullRequestRepository{db: db}
}

func (r *pullRequestRepository) CreatePR(pr domain.PullRequest) error {
	return r.db.Create(&pr).Error
}

func (r *pullRequestRepository) GetPR(id string) (domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.Preload("AssignedReviewers").First(&pr, "pull_request_id = ?", id).Error
	return pr, err
}

func (r *pullRequestRepository) GetTeamMembers(teamName string) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Where("team_name = ? AND is_active = true", teamName).Find(&users).Error
	return users, err
}

func (r *pullRequestRepository) GetUserByID(userID string) (domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "id = ?", userID).Error
	return user, err
}

func (r *pullRequestRepository) UpdatePR(pr domain.PullRequest) error {
	if err := r.db.Model(&pr).Updates(map[string]interface{}{
		"status":    pr.Status,
		"merged_at": pr.MergedAt,
	}).Error; err != nil {
		return err
	}

	if err := r.db.Where("pull_request_id = ?", pr.ID).Delete(&domain.PullRequestReviewer{}).Error; err != nil {
		return err
	}

	for _, u := range pr.AssignedReviewers {
		reviewer := domain.PullRequestReviewer{
			PullRequestID: pr.ID,
			ReviewerID:    u.ID,
		}
		if err := r.db.Create(&reviewer).Error; err != nil {
			return err
		}
	}

	return nil
}
