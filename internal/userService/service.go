package userService

import (
	"PullRequestService/domain"
	"errors"
	"gorm.io/gorm"
)

type UserService interface {
	SetIsActive(isActive bool, id string) (*domain.User, error)
	GetUserByID(id string) (domain.User, error)
	GetPRsForReviewer(userID string) ([]domain.PullRequest, error)
	DeactivateAndReassign(teamName string, userIDs []string) (*domain.DeactivateResult, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (us *userService) SetIsActive(isActive bool, id string) (*domain.User, error) {
	existing, err := us.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	existing.IsActive = isActive

	if err = us.repo.SetIsActive(existing); err != nil {
		return nil, err
	}
	return &existing, nil
}

func (us *userService) GetUserByID(id string) (domain.User, error) {
	return us.repo.GetUserByID(id)
}

func (us *userService) GetPRsForReviewer(userID string) ([]domain.PullRequest, error) {
	_, err := us.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("NOT_FOUND")
		}
		return nil, err
	}

	prs, err := us.repo.GetPRsForReviewer(userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (us *userService) DeactivateAndReassign(teamName string, userIDs []string) (*domain.DeactivateResult, error) {
	return us.repo.DeactivateAndReassign(teamName, userIDs)
}
