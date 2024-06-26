// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: users.sql

package queries

import (
	"context"
)

const addDemoUser = `-- name: AddDemoUser :one
insert into users(name, email, password_hash, password_hash_algorithm)
values ('demo user', 'demo@mahlzeit.app', '', 'argon2')
on conflict (email) do update set email = excluded.email
returning id
`

func (q *Queries) AddDemoUser(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, addDemoUser)
	var id int64
	err := row.Scan(&id)
	return id, err
}
