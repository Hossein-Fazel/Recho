package postgres_repo

import (
	"context"
	"database/sql"
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

const (
	uniqueViolation = "23505"
	undefinedColumn = "42703"

	groupModuleName = "group repo"
)

type Group struct {
	sql *sqlc.Queries
	db  *pgxpool.Pool
}

var _ application.GroupRepo = (*Group)(nil)

func NewGroupRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Group {
	pkg.Logger.Info().Msg("Initializing Group Repository")

	return &Group{
		sql: sql,
		db:  db,
	}
}

func (r *Group) Create(ctx context.Context, params application.CreateGroupParams) (*model.Group, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal(groupModuleName, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := r.sql.WithTx(tx)

	conversationID, err := q.InsertConversation(ctx, sqlc.ConversationTypeGroup)
	if err != nil {
		return nil, apperr.Internal(groupModuleName, err)
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
				groupModuleName,
				"invite code already exists",
				err,
			)
		}

		return nil, apperr.Internal(groupModuleName, err)
	}

	err = q.InsertGroupMember(ctx, sqlc.InsertGroupMemberParams{
		GroupID: conversationID,
		UserID:  params.CreatedBy,
		Role:    sqlc.GroupMemberRoleOwner,
	})
	if err != nil {
		return nil, apperr.Internal(groupModuleName, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, apperr.Internal(groupModuleName, err)
	}

	return toModelGroup(group), nil
}

func (r *Group) GetByInviteCode(ctx context.Context, code string) (*model.GroupPreview, error) {
	row, err := r.sql.GetGroupByInviteCode(ctx, pgtype.Text{
		String: code,
		Valid:  true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(groupModuleName, "invite is not valid", err)
		}

		return nil, apperr.Internal(groupModuleName, err)
	}

	return toModelGroupPreview(row), nil
}

func (r *Group) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	err := r.sql.InsertGroupMember(ctx, sqlc.InsertGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
		Role:    sqlc.GroupMemberRoleMember,
	})
	if err != nil {
		return apperr.Internal(groupModuleName, err)
	}

	return nil
}

func (r *Group) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	err := r.sql.RemoveGroupMember(ctx, sqlc.RemoveGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return apperr.Internal(groupModuleName, err)
	}

	return nil
}

func (r *Group) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	err := r.sql.DeleteGroup(ctx, groupID)
	if err != nil {
		return apperr.Internal(groupModuleName, err)
	}

	return nil
}

func (r *Group) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (model.GroupMemberRole, error) {
	role, err := r.sql.GetGroupMemberRole(ctx, sqlc.GetGroupMemberRoleParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}

		return "", apperr.Internal(groupModuleName, err)
	}

	return model.GroupMemberRole(role), nil
}

func (r *Group) UpdateInviteCode(ctx context.Context, groupID uuid.UUID, newInviteCode string) error {
	_, err := r.sql.UpdateInviteCode(ctx, sqlc.UpdateInviteCodeParams{
		InviteCode: pgtype.Text{
			String: newInviteCode,
			Valid:  true,
		},
		GroupID: groupID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.NotFound(groupModuleName, "invalid group id", err)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case undefinedColumn:
				return apperr.InvalidInput(groupModuleName, "invalid group id", err)
			case uniqueViolation:
				return apperr.Conflict(groupModuleName, "invite code already exists", err)
			}
		}

		return apperr.Internal(groupModuleName, err)
	}

	return nil
}
