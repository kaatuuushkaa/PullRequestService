package teamService

import (
	"PullRequestService/domain"
	"fmt"
	"gorm.io/gorm"
)

type TeamRepository interface {
	PostTeam(team domain.Team) error
	GetTeam(teamName string) (domain.Team, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (t *teamRepository) GetTeam(teamName string) (domain.Team, error) {
	var team domain.Team
	err := t.db.First(&team, "name = ?", teamName).Error
	if err != nil {
		fmt.Println(err)
		return domain.Team{}, err
	}
	var members []domain.User
	err = t.db.Where("team_name = ?", teamName).Find(&members).Error
	if err != nil {
		fmt.Println(err)
		return domain.Team{}, err
	}

	team.Members = members

	return team, err
}

func (t *teamRepository) PostTeam(team domain.Team) error {
	err := t.db.Create(&team).Error
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, m := range team.Members {
		err = t.db.
			Where("id = ?", m.ID).
			Assign(domain.User{
				Username: m.Username,
				TeamName: team.Name,
				IsActive: m.IsActive,
			}).
			FirstOrCreate(&m).Error
		if err != nil {
			return err
		}
	}

	return nil
}
