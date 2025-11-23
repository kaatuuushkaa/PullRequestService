package handlers

import (
	"PullRequestService/internal/userService"
	"PullRequestService/internal/web/users"
	"context"
	"errors"
)

type UserHandler struct {
	service userService.UserService
}

func NewUserHandler(service userService.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (u *UserHandler) GetUsersGetReview(ctx context.Context, request users.GetUsersGetReviewRequestObject) (users.GetUsersGetReviewResponseObject, error) {

	prs, err := u.service.GetPRsForReviewer(request.Params.UserId)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			return users.GetUsersGetReview200JSONResponse{
				UserId:       request.Params.UserId,
				PullRequests: []users.PullRequestShort{},
			}, nil
		}
		return nil, err
	}

	shortPRs := make([]users.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		shortPRs = append(shortPRs, users.PullRequestShort{
			PullRequestId:   pr.ID,
			PullRequestName: pr.Name,
			AuthorId:        pr.AuthorID,
			Status:          users.PullRequestShortStatus(pr.Status),
		})
	}

	return users.GetUsersGetReview200JSONResponse{
		UserId:       request.Params.UserId,
		PullRequests: shortPRs,
	}, nil
}

func (u *UserHandler) PostUsersSetIsActive(ctx context.Context, request users.PostUsersSetIsActiveRequestObject) (users.PostUsersSetIsActiveResponseObject, error) {
	req := request.Body
	if req == nil {
		return nil, errors.New("invalid request body")
	}

	user, err := u.service.SetIsActive(req.IsActive, req.UserId)
	if err != nil {
		return users.PostUsersSetIsActive404JSONResponse{
			Error: struct {
				Code    users.ErrorResponseErrorCode `json:"code"`
				Message string                       `json:"message"`
			}{
				Code:    users.NOTFOUND,
				Message: "resource not found",
			},
		}, nil
	}

	return users.PostUsersSetIsActive200JSONResponse{
		User: &users.User{
			UserId:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}, nil

}
