package data

import (
	"context"
	"database/sql"
	"time"
)

type Permissions []string

type PermissionModel struct {
	DB *sql.DB
}

func (perms Permissions) Include(code string) bool {
	for _, p := range perms {
		if code == p {
			return true
		}
	}
	return false
}

func (m *PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `select p.code from permissions p
	inner join users_permissions up on p.id = up.permission_id
	inner join users u on u.id = up.user_id
	where up.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqlRows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var permissions Permissions

	defer sqlRows.Close()

	for sqlRows.Next() {
		var permission string

		err = sqlRows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = sqlRows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
