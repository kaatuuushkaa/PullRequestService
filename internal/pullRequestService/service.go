package pullRequestService

import (
	"PullRequestService/domain"
	"errors"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type PullRequestService interface {
	CreatePR(prID, prName, authorID string) (domain.PullRequest, error)
	MergePR(prID string) (domain.PullRequest, error)
	ReassignReviewer(prID, oldUserID string) (domain.PullRequest, string, error)
}

type pullRequestService struct {
	repo PullRequestRepository
}

func NewPullRequestService(repo PullRequestRepository) PullRequestService {
	return &pullRequestService{
		repo: repo,
	}
}

func (s *pullRequestService) CreatePR(prID, prName, authorID string) (domain.PullRequest, error) {
	if _, err := s.repo.GetPR(prID); err == nil {
		return domain.PullRequest{}, errors.New("PR_EXISTS")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.PullRequest{}, err
	}

	author, err := s.repo.GetUserByID(authorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PullRequest{}, errors.New("NOT_FOUND")
		}
		return domain.PullRequest{}, err
	}

	candidates, err := s.repo.GetTeamMembers(author.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	var reviewers []domain.User
	for _, u := range candidates {
		if u.ID != author.ID {
			reviewers = append(reviewers, u)
		}
	}

	rand.Seed(time.Now().UnixNano())
	if len(reviewers) > 2 {
		rand.Shuffle(len(reviewers), func(i, j int) { reviewers[i], reviewers[j] = reviewers[j], reviewers[i] })
		reviewers = reviewers[:2]
	}

	pr := domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          author.ID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreatePR(pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *pullRequestService) MergePR(prID string) (domain.PullRequest, error) {
	pr, err := s.repo.GetPR(prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PullRequest{}, errors.New("NOT_FOUND")
		}
		return domain.PullRequest{}, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	now := time.Now()
	pr.Status = "MERGED"
	pr.MergedAt = &now

	if err := s.repo.UpdatePR(pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *pullRequestService) ReassignReviewer(prID, oldUserID string) (domain.PullRequest, string, error) {
	pr, err := s.repo.GetPR(prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.PullRequest{}, "", errors.New("PR_NOT_FOUND")
		}
		return domain.PullRequest{}, "", err
	}

	if pr.Status == "MERGED" {
		return domain.PullRequest{}, "", errors.New("PR_MERGED")
	}

	var index int = -1
	for i, u := range pr.AssignedReviewers {
		if u.ID == oldUserID {
			index = i
			break
		}
	}
	if index == -1 {
		return domain.PullRequest{}, "", errors.New("NOT_ASSIGNED")
	}

	oldUser, err := s.repo.GetUserByID(oldUserID)
	if err != nil {
		return domain.PullRequest{}, "", errors.New("USER_NOT_FOUND")
	}

	candidates, err := s.repo.GetTeamMembers(oldUser.TeamName)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	var available []domain.User
	for _, u := range candidates {
		if u.ID != oldUserID {
			skip := false
			for _, r := range pr.AssignedReviewers {
				if r.ID == u.ID {
					skip = true
					break
				}
			}
			if !skip {
				available = append(available, u)
			}
		}
	}

	if len(available) == 0 {
		return domain.PullRequest{}, "", errors.New("NO_CANDIDATE")
	}

	rand.Seed(time.Now().UnixNano())
	newReviewer := available[rand.Intn(len(available))]
	pr.AssignedReviewers[index] = newReviewer

	if err := s.repo.UpdatePR(pr); err != nil {
		return domain.PullRequest{}, "", err
	}

	return pr, newReviewer.ID, nil
}
