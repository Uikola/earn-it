package task

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	executor := postgres.GetQueryExecutor(ctx, r.pool)
	return sqlc.New(executor)
}

func (r *Repository) CreateTask(ctx context.Context, userID int64, title string, scheduledDate time.Time, rewardValue int32) (models.Task, error) {
	q := r.queries(ctx)

	task, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		UserID:        userID,
		Title:         title,
		ScheduledDate: pgtype.Date{Time: scheduledDate, Valid: true},
		RewardValue:   rewardValue,
	})
	if err != nil {
		return models.Task{}, err
	}

	return toDomainTask(task), nil
}

func (r *Repository) TaskByID(ctx context.Context, taskID int64) (models.Task, error) {
	q := r.queries(ctx)

	task, err := q.TaskByID(ctx, taskID)
	if err != nil {
		return models.Task{}, err
	}

	return toDomainTask(task), nil
}

func (r *Repository) TasksByUserAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Task, error) {
	q := r.queries(ctx)

	tasks, err := q.TasksByUserAndDateRange(ctx, sqlc.TasksByUserAndDateRangeParams{
		UserID:          userID,
		ScheduledDate:   pgtype.Date{Time: from, Valid: true},
		ScheduledDate_2: pgtype.Date{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	domainTasks := make([]models.Task, 0, len(tasks))
	for _, t := range tasks {
		domainTasks = append(domainTasks, toDomainTask(t))
	}

	return domainTasks, nil
}

func (r *Repository) UpdateTask(ctx context.Context, taskID int64, task models.Task) error {
	q := r.queries(ctx)

	var completedAt pgtype.Timestamptz
	if task.CompletedAt != nil {
		completedAt = pgtype.Timestamptz{Time: *task.CompletedAt, Valid: true}
	}

	return q.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:            taskID,
		Title:         task.Title,
		ScheduledDate: pgtype.Date{Time: task.ScheduledDate, Valid: true},
		Status:        task.Status,
		CompletedAt:   completedAt,
	})
}

func (r *Repository) RescheduleExpiredTasks(ctx context.Context, userID int64, today time.Time) error {
	q := r.queries(ctx)

	return q.RescheduleExpiredTasks(ctx, sqlc.RescheduleExpiredTasksParams{
		UserID:        userID,
		ScheduledDate: pgtype.Date{Time: today, Valid: true},
	})
}

func (r *Repository) DeleteTask(ctx context.Context, taskID int64) error {
	q := r.queries(ctx)
	return q.DeleteTask(ctx, taskID)
}

func toDomainTask(t sqlc.Task) models.Task {
	var projectID *int64
	if t.ProjectID.Valid {
		projectID = &t.ProjectID.Int64
	}

	var completedAt *time.Time
	if t.CompletedAt.Valid {
		completedAt = &t.CompletedAt.Time
	}

	return models.Task{
		ID:            t.ID,
		UserID:        t.UserID,
		ProjectID:     projectID,
		Title:         t.Title,
		ScheduledDate: t.ScheduledDate.Time,
		RewardValue:   t.RewardValue,
		Status:        t.Status,
		CompletedAt:   completedAt,
		CreatedAt:     t.CreatedAt.Time,
	}
}
