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

func (s *userService) GetPRsForReviewer(userID string) ([]domain.PullRequest, error) {
	_, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("NOT_FOUND")
		}
		return nil, err
	}

	prs, err := s.repo.GetPRsForReviewer(userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
