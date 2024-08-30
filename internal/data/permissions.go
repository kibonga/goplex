package data

import (
	"context"
	"database/sql"
	"time"
)

type Permission struct {
	ID   int64
	Code string
}

type PermissionModel struct {
	DB sql.DB
}

func (m *PermissionModel) GetAllForUser(userID int64) ([]*Permission, error) {
	query := `select p.id, p.code from permissions p
	inner join users_permissions up on p.id = up.permission_id
	inner join users u on u.id = up.user_id
	where up.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqlRows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	permissions := []*Permission{}

	var permission Permission

	defer sqlRows.Close()

	for sqlRows.Next() {
		err = sqlRows.Scan(
			&permission.ID,
			&permission.Code,
		)

		if err != nil {
			return nil, err
		}

		permissions = append(permissions, &permission)
	}

	return permissions, nil
}
