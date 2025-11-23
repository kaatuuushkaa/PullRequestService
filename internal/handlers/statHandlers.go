package handlers

import (
	"PullRequestService/internal/statsService"
	"PullRequestService/internal/web/stats"
	"context"
)

type StatsHandler struct {
	service statsService.StatsService
}

func NewStatsHandler(service statsService.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetStats(ctx context.Context, request stats.GetStatsRequestObject) (stats.GetStatsResponseObject, error) {
	result, err := h.service.GetStats()
	if err != nil {
		return stats.GetStats500JSONResponse{
			Error: struct {
				Code    stats.ErrorResponseErrorCode `json:"code"`
				Message string                       `json:"message"`
			}{
				Code:    stats.NOTFOUND,
				Message: err.Error(),
			},
		}, nil
	}

	return stats.GetStats200JSONResponse(result), nil
}
