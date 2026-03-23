package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/api"
)

type StrictCalendarHandler struct {
	app *app.App
}

func NewStrictCalendarHandler(app *app.App) *StrictCalendarHandler {
	return &StrictCalendarHandler{
		app: app,
	}
}

var _ api.StrictServerInterface = (*StrictCalendarHandler)(nil)

func (h *StrictCalendarHandler) CreateEvent(
	ctx context.Context,
	request api.CreateEventRequestObject,
) (api.CreateEventResponseObject, error) {
	if request.Body == nil {
		return api.CreateEvent400JSONResponse{
			Code:    400,
			Message: "Empty request body",
		}, nil
	}

	createdEvent, err := h.app.CreateEvent(ctx, request.Body.Duration, request.Body.Description,
		request.Body.NotifyBefore, request.Body.EventTime, request.Body.Title, request.Body.UserId)
	if err != nil {
		if errors.Is(err, model.ErrEventAlreadyExists) {
			return api.CreateEvent400JSONResponse{
				Code:    404,
				Message: "Event already exists",
			}, nil
		}
		return api.CreateEvent500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.CreateEvent201JSONResponse(h.convertToAPIEvent(createdEvent)), nil
}

func (h *StrictCalendarHandler) GetEventById(
	ctx context.Context,
	request api.GetEventByIdRequestObject,
) (api.GetEventByIdResponseObject, error) {
	eventID := uuid.UUID(request.Id)

	event, err := h.app.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, model.ErrEventNotFound) {
			return api.GetEventById404JSONResponse{
				Code:    404,
				Message: "Event not found",
			}, nil
		}
		return api.GetEventById500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.GetEventById200JSONResponse(h.convertToAPIEvent(event)), nil
}

func (h *StrictCalendarHandler) UpdateEvent(
	ctx context.Context,
	request api.UpdateEventRequestObject,
) (api.UpdateEventResponseObject, error) {
	if request.Body == nil {
		return api.UpdateEvent400JSONResponse{
			Code:    400,
			Message: "Empty request body",
		}, nil
	}

	eventID := uuid.UUID(request.Id)

	updatedEvent, err := h.app.UpdateEvent(
		ctx, eventID, request.Body.Description, request.Body.Duration,
		request.Body.NotifyBefore, request.Body.Title, request.Body.EventTime, request.Body.UserId)
	if err != nil {
		if errors.Is(err, model.ErrEventNotFound) {
			return api.UpdateEvent400JSONResponse{
				Code:    404,
				Message: "Event not found",
			}, nil
		}
		return api.UpdateEvent500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.UpdateEvent200JSONResponse(h.convertToAPIEvent(updatedEvent)), nil
}

func (h *StrictCalendarHandler) DeleteEvent(
	ctx context.Context,
	request api.DeleteEventRequestObject,
) (api.DeleteEventResponseObject, error) {
	eventID := uuid.UUID(request.Id)

	err := h.app.Delete(ctx, eventID)
	if err != nil {
		if errors.Is(err, model.ErrEventNotFound) {
			return api.DeleteEvent404JSONResponse{
				Code:    404,
				Message: "Event not found",
			}, nil
		}
		return api.DeleteEvent500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.DeleteEvent204Response{}, nil
}

func (h *StrictCalendarHandler) ListEventsByDay(
	ctx context.Context,
	request api.ListEventsByDayRequestObject,
) (api.ListEventsByDayResponseObject, error) {
	date := request.Params.Date.Time

	events, err := h.app.ListByDay(ctx, date)
	if err != nil {
		return api.ListEventsByDay500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.ListEventsByDay200JSONResponse(h.convertToAPIEventList(events)), nil
}

func (h *StrictCalendarHandler) ListEventsByWeek(
	ctx context.Context,
	request api.ListEventsByWeekRequestObject,
) (api.ListEventsByWeekResponseObject, error) {
	date := request.Date.Time

	events, err := h.app.ListByWeek(ctx, date)
	if err != nil {
		return api.ListEventsByWeek500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.ListEventsByWeek200JSONResponse(h.convertToAPIEventList(events)), nil
}

func (h *StrictCalendarHandler) ListEventsByMonth(
	ctx context.Context,
	request api.ListEventsByMonthRequestObject,
) (api.ListEventsByMonthResponseObject, error) {
	date := request.Date.Time

	events, err := h.app.ListByMonth(ctx, date)
	if err != nil {
		return api.ListEventsByMonth500JSONResponse{
			Code:    500,
			Message: "Internal server error",
		}, nil
	}

	return api.ListEventsByMonth200JSONResponse(h.convertToAPIEventList(events)), nil
}

func (h *StrictCalendarHandler) convertToAPIEvent(event model.Event) api.Event {
	result := api.Event{
		Id:        openapi_types.UUID(event.ID),
		Title:     event.Title,
		EventTime: event.EventTime,
		UserId:    openapi_types.UUID(event.UserID),
		Duration:  event.Duration.String(),
	}

	if event.Description != "" {
		result.Description = &event.Description
	}

	if event.NotifyBefore != 0 {
		notifyBeforeStr := event.NotifyBefore.String()
		result.NotifyBefore = &notifyBeforeStr
	}

	return result
}

func (h *StrictCalendarHandler) convertToAPIEventList(events []model.Event) []api.Event {
	result := make([]api.Event, len(events))
	for i, event := range events {
		result[i] = h.convertToAPIEvent(event)
	}
	return result
}
