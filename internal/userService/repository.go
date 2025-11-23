package userService

import (
	"PullRequestService/domain"
	"errors"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(id string) (domain.User, error)
	SetIsActive(user domain.User) error
	GetPRsForReviewer(userID string) ([]domain.PullRequest, error)
	DeactivateAndReassign(teamName string, userIDs []string) (*domain.DeactivateResult, error)
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

func (u *userRepository) GetPRsForReviewer(userID string) ([]domain.PullRequest, error) {
	var prs []domain.PullRequest
	err := u.db.Joins("JOIN pull_request_reviewers prr ON prr.pull_request_id = pull_requests.pull_request_id").
		Where("prr.reviewer_id = ?", userID).
		Find(&prs).Error
	return prs, err
}

func (u *userRepository) DeactivateAndReassign(teamName string, userIDs []string) (*domain.DeactivateResult, error) {
	result := &domain.DeactivateResult{
		TeamName: teamName,
	}

	return result, u.db.Transaction(func(tx *gorm.DB) error {

		var teamCount int64
		if err := tx.Table("teams").
			Where("name = ?", teamName).
			Count(&teamCount).Error; err != nil {
			return err
		}
		if teamCount == 0 {
			return errors.New("TEAM_NOT_FOUND")
		}

		if len(userIDs) == 0 {
			return nil
		}

		var deactivated []domain.User
		if err := tx.Model(&domain.User{}).
			Where("team_name = ? AND id IN ?", teamName, userIDs).
			Updates(map[string]any{"is_active": false}).
			Error; err != nil {
			return err
		}

		if err := tx.Where("team_name = ? AND id IN ?", teamName, userIDs).
			Find(&deactivated).Error; err != nil {
			return err
		}
		result.DeactivatedCount = len(deactivated)

		if result.DeactivatedCount == 0 {
			return nil
		}

		var candidates []domain.User
		if err := tx.Where("team_name = ? AND is_active = true", teamName).
			Find(&candidates).Error; err != nil {
			return err
		}

		activeCandidates := make([]string, 0, len(candidates))
		for _, u := range candidates {
			activeCandidates = append(activeCandidates, u.ID)
		}

		type Row struct {
			PRID     string
			AuthorID string
			Reviewer string
		}

		var rows []Row
		if err := tx.Table("pull_requests as pr").
			Select("pr.pull_request_id, pr.author_id, prr.reviewer_id").
			Joins("JOIN pull_request_reviewers prr ON prr.pull_request_id = pr.pull_request_id").
			Where("pr.status = 'OPEN' AND prr.reviewer_id IN ?", userIDs).
			Scan(&rows).Error; err != nil {
			return err
		}

		if len(rows) == 0 {
			return nil
		}

		result.AffectedPRCount = len(rows)

		var prIDs []string
		prSet := map[string]struct{}{}
		for _, r := range rows {
			if _, ok := prSet[r.PRID]; !ok {
				prSet[r.PRID] = struct{}{}
				prIDs = append(prIDs, r.PRID)
			}
		}

		type ReviewerRow struct {
			PRID     string
			Reviewer string
		}

		var revRows []ReviewerRow
		if err := tx.Table("pull_request_reviewers").
			Where("pull_request_id IN ?", prIDs).
			Scan(&revRows).Error; err != nil {
			return err
		}

		reviewers := make(map[string]map[string]struct{})
		for _, r := range revRows {
			if reviewers[r.PRID] == nil {
				reviewers[r.PRID] = make(map[string]struct{})
			}
			reviewers[r.PRID][r.Reviewer] = struct{}{}
		}

		pickCandidate := func(author string, current map[string]struct{}) (string, bool) {
			for _, c := range activeCandidates {
				if c == author {
					continue
				}
				if _, used := current[c]; used {
					continue
				}
				return c, true
			}
			return "", false
		}

		for _, row := range rows {
			cur := reviewers[row.PRID]
			newID, ok := pickCandidate(row.AuthorID, cur)

			if ok {
				if err := tx.Table("pull_request_reviewers").
					Where("pull_request_id = ? AND reviewer_id = ?", row.PRID, row.Reviewer).
					Update("reviewer_id", newID).Error; err != nil {
					return err
				}
				delete(cur, row.Reviewer)
				cur[newID] = struct{}{}
				result.ReassignedReviewersCount++

			} else {
				if err := tx.Table("pull_request_reviewers").
					Where("pull_request_id = ? AND reviewer_id = ?", row.PRID, row.Reviewer).
					Delete(nil).Error; err != nil {
					return err
				}
				delete(cur, row.Reviewer)
			}
		}

		return nil
	})
}
