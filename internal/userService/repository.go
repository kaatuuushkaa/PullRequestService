package userService

import (
	"PullRequestService/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(id string) (domain.User, error)
	SetIsActive(user domain.User) error
	GetPRsForReviewer(userID string) ([]domain.PullRequest, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (u *userRepository) GetUserByID(id string) (domain.User, error) {
	var user domain.User
	err := u.db.First(&user, "id = ?", id).Error
	return user, err
}

func (u *userRepository) SetIsActive(user domain.User) error {
	return u.db.Save(&user).Error
}

func (r *userRepository) GetPRsForReviewer(userID string) ([]domain.PullRequest, error) {
	var prs []domain.PullRequest
	err := r.db.Joins("JOIN pull_request_reviewers prr ON prr.pull_request_id = pull_requests.pull_request_id").
		Where("prr.reviewer_id = ?", userID).
		Find(&prs).Error
	return prs, err
}
