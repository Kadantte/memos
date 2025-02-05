package mysql

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	storepb "github.com/usememos/memos/proto/gen/store"
	"github.com/usememos/memos/store"
)

func (d *DB) CreateIdentityProvider(ctx context.Context, create *store.IdentityProvider) (*store.IdentityProvider, error) {
	placeholders := []string{"?", "?", "?", "?"}
	fields := []string{"`name`", "`type`", "`identifier_filter`", "`config`"}
	args := []any{create.Name, create.Type.String(), create.IdentifierFilter, create.Config}

	stmt := "INSERT INTO `idp` (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	create.ID = int32(id)
	return create, nil
}

func (d *DB) ListIdentityProviders(ctx context.Context, find *store.FindIdentityProvider) ([]*store.IdentityProvider, error) {
	where, args := []string{"1 = 1"}, []any{}
	if v := find.ID; v != nil {
		where, args = append(where, "`id` = ?"), append(args, *v)
	}

	rows, err := d.db.QueryContext(ctx, "SELECT `id`, `name`, `type`, `identifier_filter`, `config` FROM `idp` WHERE "+strings.Join(where, " AND ")+" ORDER BY `id` ASC",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var identityProviders []*store.IdentityProvider
	for rows.Next() {
		var identityProvider store.IdentityProvider
		var typeString string
		if err := rows.Scan(
			&identityProvider.ID,
			&identityProvider.Name,
			&typeString,
			&identityProvider.IdentifierFilter,
			&identityProvider.Config,
		); err != nil {
			return nil, err
		}
		identityProvider.Type = storepb.IdentityProvider_Type(storepb.IdentityProvider_Type_value[typeString])
		identityProviders = append(identityProviders, &identityProvider)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return identityProviders, nil
}

func (d *DB) GetIdentityProvider(ctx context.Context, find *store.FindIdentityProvider) (*store.IdentityProvider, error) {
	list, err := d.ListIdentityProviders(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	identityProvider := list[0]
	return identityProvider, nil
}

func (d *DB) UpdateIdentityProvider(ctx context.Context, update *store.UpdateIdentityProvider) (*store.IdentityProvider, error) {
	set, args := []string{}, []any{}
	if v := update.Name; v != nil {
		set, args = append(set, "`name` = ?"), append(args, *v)
	}
	if v := update.IdentifierFilter; v != nil {
		set, args = append(set, "`identifier_filter` = ?"), append(args, *v)
	}
	if v := update.Config; v != nil {
		set, args = append(set, "`config` = ?"), append(args, *v)
	}
	args = append(args, update.ID)

	stmt := "UPDATE `idp` SET " + strings.Join(set, ", ") + " WHERE `id` = ?"
	_, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	identityProvider, err := d.GetIdentityProvider(ctx, &store.FindIdentityProvider{
		ID: &update.ID,
	})
	if err != nil {
		return nil, err
	}
	if identityProvider == nil {
		return nil, errors.Errorf("idp %d not found", update.ID)
	}
	return identityProvider, nil
}

func (d *DB) DeleteIdentityProvider(ctx context.Context, delete *store.DeleteIdentityProvider) error {
	where, args := []string{"`id` = ?"}, []any{delete.ID}
	stmt := "DELETE FROM `idp` WHERE " + strings.Join(where, " AND ")
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}
	if _, err = result.RowsAffected(); err != nil {
		return err
	}
	return nil
}
