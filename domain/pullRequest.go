package domain

import (
	"time"
)

type PullRequest struct {
	ID                string     `gorm:"column:pull_request_id;primaryKey"`
	Name              string     `gorm:"column:pull_request_name"`
	AuthorID          string     `gorm:"column:author_id"`
	Status            string     `gorm:"column:status"`
	AssignedReviewers []User     `gorm:"many2many:pull_request_reviewers;joinForeignKey:PullRequestID;joinReferences:ReviewerID"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	MergedAt          *time.Time `gorm:"column:merged_at"`
}

type PullRequestReviewer struct {
	PullRequestID string `gorm:"column:pull_request_id;primaryKey"`
	ReviewerID    string `gorm:"column:reviewer_id;primaryKey"`
}
