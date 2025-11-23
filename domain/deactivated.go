package domain

type DeactivateResult struct {
	TeamName                 string
	DeactivatedCount         int
	AffectedPRCount          int
	ReassignedReviewersCount int
}
