package handlers

import (
	"PullRequestService/internal/pullRequestService"
	"PullRequestService/internal/web/pullRequests"
	"context"
	"errors"
)

type PullRequestHandler struct {
	service pullRequestService.PullRequestService
}

func NewPullRequestHandler(service pullRequestService.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		service: service,
	}
}

func (p *PullRequestHandler) PostPullRequestCreate(ctx context.Context, request pullRequests.PostPullRequestCreateRequestObject) (pullRequests.PostPullRequestCreateResponseObject, error) {
	req := request.Body
	if req == nil {
		return nil, errors.New("invalid request body")
	}

	pr, err := p.service.CreatePR(req.PullRequestId, req.PullRequestName, req.AuthorId)
	if err != nil {
		switch err.Error() {
		case "PR_EXISTS":
			return pullRequests.PostPullRequestCreate409JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.PREXISTS,
					Message: "PR id already exists",
				},
			}, nil
		case "NOT_FOUND":
			return pullRequests.PostPullRequestCreate404JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.NOTFOUND,
					Message: "author not found",
				},
			}, nil
		default:
			return nil, err
		}
	}

	reviewerIDs := make([]string, len(pr.AssignedReviewers))
	for i, u := range pr.AssignedReviewers {
		reviewerIDs[i] = u.ID
	}

	return pullRequests.PostPullRequestCreate201JSONResponse{
		Pr: &pullRequests.PullRequest{
			PullRequestId:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorId:          pr.AuthorID,
			Status:            pullRequests.OPEN,
			AssignedReviewers: reviewerIDs,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          nil,
		},
	}, nil
}

func (p *PullRequestHandler) PostPullRequestMerge(ctx context.Context, request pullRequests.PostPullRequestMergeRequestObject) (pullRequests.PostPullRequestMergeResponseObject, error) {
	req := request.Body
	if req == nil {
		return nil, errors.New("invalid request body")
	}

	pr, err := p.service.MergePR(req.PullRequestId)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			return pullRequests.PostPullRequestMerge404JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.NOTFOUND,
					Message: "PR not found",
				},
			}, nil
		}
		return nil, err
	}

	reviewerIDs := make([]string, len(pr.AssignedReviewers))
	for i, u := range pr.AssignedReviewers {
		reviewerIDs[i] = u.ID
	}

	return pullRequests.PostPullRequestMerge200JSONResponse{
		Pr: &pullRequests.PullRequest{
			PullRequestId:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorId:          pr.AuthorID,
			Status:            pullRequests.MERGED,
			AssignedReviewers: reviewerIDs,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          pr.MergedAt,
		},
	}, nil
}

func (p *PullRequestHandler) PostPullRequestReassign(ctx context.Context, request pullRequests.PostPullRequestReassignRequestObject) (pullRequests.PostPullRequestReassignResponseObject, error) {
	req := request.Body
	if req == nil {
		return nil, errors.New("invalid request body")
	}

	newPR, replacedBy, err := p.service.ReassignReviewer(req.PullRequestId, req.OldUserId)
	if err != nil {
		switch err.Error() {
		case "PR_NOT_FOUND", "USER_NOT_FOUND":
			return pullRequests.PostPullRequestReassign404JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.NOTFOUND,
					Message: "PR or user not found",
				},
			}, nil
		case "PR_MERGED":
			return pullRequests.PostPullRequestReassign409JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.PRMERGED,
					Message: "cannot reassign on merged PR",
				},
			}, nil
		case "NOT_ASSIGNED":
			return pullRequests.PostPullRequestReassign409JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.NOTASSIGNED,
					Message: "reviewer is not assigned to this PR",
				},
			}, nil
		case "NO_CANDIDATE":
			return pullRequests.PostPullRequestReassign409JSONResponse{
				Error: struct {
					Code    pullRequests.ErrorResponseErrorCode `json:"code"`
					Message string                              `json:"message"`
				}{
					Code:    pullRequests.NOCANDIDATE,
					Message: "no active replacement candidate in team",
				},
			}, nil
		default:
			return nil, err
		}
	}

	reviewerIDs := make([]string, len(newPR.AssignedReviewers))
	for i, u := range newPR.AssignedReviewers {
		reviewerIDs[i] = u.ID
	}

	return pullRequests.PostPullRequestReassign200JSONResponse{
		Pr: pullRequests.PullRequest{
			PullRequestId:     newPR.ID,
			PullRequestName:   newPR.Name,
			AuthorId:          newPR.AuthorID,
			Status:            pullRequests.PullRequestStatus(newPR.Status),
			AssignedReviewers: reviewerIDs,
			CreatedAt:         &newPR.CreatedAt,
			MergedAt:          newPR.MergedAt,
		},
		ReplacedBy: replacedBy,
	}, nil
}
