package postgres_repo

import (
	"context"
	"errors"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type Group struct {
	sql *sqlc.Queries
	db  *pgxpool.Pool
}

func NewGroupRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Group {
	pkg.Logger.Info().Msg("Initializing Group Repository")

	return &Group{
		sql: sql,
		db:  db,
	}
}

func (r *Group) Create(ctx context.Context, params application.CreateGroupParams) (*model.Group, error) {
	if params.CreatedBy == uuid.Nil {
		return nil, apperr.InvalidInput("group repo", "creator id is required", nil)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal("group repo", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := r.sql.WithTx(tx)

	conversationID, err := q.InsertConversation(ctx, sqlc.ConversationTypeGroup)
	if err != nil {
		return nil, apperr.Internal("group repo", err)
	}

	group, err := q.InsertGroup(ctx, sqlc.InsertGroupParams{
		ConversationID: conversationID,
		Name:           params.Name,
		Bio:            pgtype.Text{String: params.Bio, Valid: params.Bio != ""},
		InviteCode: pgtype.Text{
			String: params.InviteCode,
			Valid:  params.InviteCode != "",
		},
		CreatedBy: params.CreatedBy,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return nil, apperr.Conflict(
				"group repo",
				"invite code already exists",
				err,
			)
		}

		return nil, apperr.Internal("group repo", err)
	}

	err = q.InsertGroupMember(ctx, sqlc.InsertGroupMemberParams{
		GroupID: conversationID,
		UserID:  params.CreatedBy,
		Role:    sqlc.GroupMemberRoleOwner,
	})
	if err != nil {
		return nil, apperr.Internal("group repo", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, apperr.Internal("group repo", err)
	}

	return toModelGroup(group), nil
}

func (r *Group) GetByInviteCode(ctx context.Context, code string) (*model.GroupPreview, error) {
	if code == "" {
		return nil, apperr.InvalidInput("group repo", "invite code is required", nil)
	}

	row, err := r.sql.GetGroupByInviteCode(ctx, pgtype.Text{
		String: code,
		Valid:  true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound("group repo", "invite is not valid", err)
		}

		return nil, apperr.Internal("group repo", err)
	}

	return toModelGroupPreview(row), nil
}

func (r *Group) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	if groupID == uuid.Nil || userID == uuid.Nil {
		return apperr.InvalidInput("group repo", "group and user id are required", nil)
	}

	err := r.sql.InsertGroupMember(ctx, sqlc.InsertGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
		Role:    sqlc.GroupMemberRoleMember,
	})
	if err != nil {
		return apperr.Internal("group repo", err)
	}

	return nil
}

func (r *Group) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	if groupID == uuid.Nil || userID == uuid.Nil {
		return apperr.InvalidInput("group repo", "group and user id are required", nil)
	}

	err := r.sql.RemoveGroupMember(ctx, sqlc.RemoveGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return apperr.Internal("group repo", err)
	}

	return nil
}

func (r *Group) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	if groupID == uuid.Nil {
		return apperr.InvalidInput("group repo", "group id is required", nil)
	}

	err := r.sql.DeleteGroup(ctx, groupID)
	if err != nil {
		return apperr.Internal("group repo", err)
	}

	return nil
}

func (r *Group) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (model.GroupMemberRole, error) {
	if groupID == uuid.Nil || userID == uuid.Nil {
		return "", apperr.InvalidInput("group repo", "group and user id are required", nil)
	}

	role, err := r.sql.GetGroupMemberRole(ctx, sqlc.GetGroupMemberRoleParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}

		return "", apperr.Internal("group repo", err)
	}

	return model.GroupMemberRole(role), nil
}
