package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/model"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/api"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/handlers"
	memorystorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func createTestEvent() model.Event {
	return model.Event{
		ID:           uuid.New(),
		Title:        "Test Event",
		EventTime:    time.Now().Truncate(time.Second), // убираем наносекунды
		Duration:     time.Hour,
		Description:  "Test Description",
		UserID:       uuid.New(),
		NotifyBefore: 30 * time.Minute,
	}
}

func TestCreateEvent(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	event := createTestEvent()
	createReq := api.CreateEventRequestObject{
		Body: &api.CreateEventRequest{
			Title:        event.Title,
			EventTime:    event.EventTime,
			Duration:     event.Duration.String(),
			Description:  &event.Description,
			UserId:       openapi_types.UUID(event.UserID),
			NotifyBefore: strPtr(event.NotifyBefore.String()),
		},
	}

	resp, err := handler.CreateEvent(context.Background(), createReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	createResp, ok := resp.(api.CreateEvent201JSONResponse)
	require.True(t, ok, "expected 201 response")

	assert.Equal(t, event.Title, createResp.Title)
	assert.Equal(t, event.EventTime, createResp.EventTime)
	assert.Equal(t, event.Description, *createResp.Description)
	assert.Equal(t, openapi_types.UUID(event.UserID), createResp.UserId)
}

func TestGetEventById(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	event := createTestEvent()
	_, err := store.Create(context.Background(), event)
	require.NoError(t, err)

	getReq := api.GetEventByIdRequestObject{
		Id: openapi_types.UUID(event.ID),
	}

	resp, err := handler.GetEventById(context.Background(), getReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	getResp, ok := resp.(api.GetEventById200JSONResponse)
	require.True(t, ok, "expected 200 response")

	assert.Equal(t, openapi_types.UUID(event.ID), getResp.Id)
	assert.Equal(t, event.Title, getResp.Title)
}

func TestGetEventById_NotFound(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	getReq := api.GetEventByIdRequestObject{
		Id: openapi_types.UUID(uuid.New()),
	}

	resp, err := handler.GetEventById(context.Background(), getReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	notFoundResp, ok := resp.(api.GetEventById404JSONResponse)
	require.True(t, ok, "expected 404 response")
	assert.Equal(t, 404, int(notFoundResp.Code))
	assert.Contains(t, notFoundResp.Message, "not found")
}

func TestUpdateEvent(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	event := createTestEvent()
	_, err := store.Create(context.Background(), event)
	require.NoError(t, err)

	newTitle := "Updated Title"
	newDuration := 2 * time.Hour

	updateReq := api.UpdateEventRequestObject{
		Id: openapi_types.UUID(event.ID),
		Body: &api.UpdateEventRequest{
			Title:    &newTitle,
			Duration: strPtr(newDuration.String()),
		},
	}

	resp, err := handler.UpdateEvent(context.Background(), updateReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	updateResp, ok := resp.(api.UpdateEvent200JSONResponse)
	require.True(t, ok, "expected 200 response")

	assert.Equal(t, newTitle, updateResp.Title)
	assert.Equal(t, newDuration.String(), updateResp.Duration)

	assert.Equal(t, event.EventTime, updateResp.EventTime)
}

func TestDeleteEvent(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	event := createTestEvent()
	_, err := store.Create(context.Background(), event)
	require.NoError(t, err)

	deleteReq := api.DeleteEventRequestObject{
		Id: openapi_types.UUID(event.ID),
	}

	resp, err := handler.DeleteEvent(context.Background(), deleteReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	_, ok := resp.(api.DeleteEvent204Response)
	require.True(t, ok, "expected 204 response")

	_, err = store.GetByID(context.Background(), event.ID)
	assert.ErrorIs(t, err, model.ErrEventNotFound)
}

func TestListEventsByDay(t *testing.T) {
	logg := logger.NewNoopLogger()
	store := memorystorage.New()
	calendar := app.New(logg, store)
	handler := handlers.NewStrictCalendarHandler(calendar)

	today := time.Now().Truncate(24 * time.Hour)

	for i := 0; i < 3; i++ {
		event := model.Event{
			ID:        uuid.New(),
			Title:     "Event",
			EventTime: today.Add(time.Duration(i) * time.Hour),
			UserID:    uuid.New(),
		}
		_, err := store.Create(context.Background(), event)
		require.NoError(t, err)
	}

	otherDay := today.Add(24 * time.Hour)
	otherEvent := model.Event{
		ID:        uuid.New(),
		Title:     "Other Day Event",
		EventTime: otherDay,
		UserID:    uuid.New(),
	}
	_, err := store.Create(context.Background(), otherEvent)
	require.NoError(t, err)

	listReq := api.ListEventsByDayRequestObject{
		Params: api.ListEventsByDayParams{
			Date: openapi_types.Date{Time: today},
		},
	}

	resp, err := handler.ListEventsByDay(context.Background(), listReq)

	require.NoError(t, err)
	require.NotNil(t, resp)

	listResp, ok := resp.(api.ListEventsByDay200JSONResponse)
	require.True(t, ok, "expected 200 response")

	assert.Len(t, listResp, 3)
}

func strPtr(s string) *string {
	return &s
}
