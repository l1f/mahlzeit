// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: ingredients.sql

package queries

import (
	"context"

	"github.com/jackc/pgtype"
)

const addIngredient = `-- name: AddIngredient :one
insert into ingredients (name)
values ($1)
on conflict (name) do update set name=excluded.name -- no-op that effectively does nothing, but returns the ID as intended
returning id
`

func (q *Queries) AddIngredient(ctx context.Context, name string) (int64, error) {
	row := q.db.QueryRow(ctx, addIngredient, name)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const addIngredientToStep = `-- name: AddIngredientToStep :exec
insert into step_ingredients (step_id, ingredients_id, unit_id, amount, note)
values ($1,
		$2,
		nullif($3::bigint, 0),
		$4,
		$5)
`

type AddIngredientToStepParams struct {
	StepID        int64
	IngredientsID int64
	UnitID        int64
	Amount        pgtype.Numeric
	Note          string
}

func (q *Queries) AddIngredientToStep(ctx context.Context, arg AddIngredientToStepParams) error {
	_, err := q.db.Exec(ctx, addIngredientToStep,
		arg.StepID,
		arg.IngredientsID,
		arg.UnitID,
		arg.Amount,
		arg.Note,
	)
	return err
}

const deleteIngredientFromStep = `-- name: DeleteIngredientFromStep :exec
delete
from step_ingredients
where step_id = $1
  and ingredients_id = $2
`

type DeleteIngredientFromStepParams struct {
	StepID        int64
	IngredientsID int64
}

func (q *Queries) DeleteIngredientFromStep(ctx context.Context, arg DeleteIngredientFromStepParams) error {
	_, err := q.db.Exec(ctx, deleteIngredientFromStep, arg.StepID, arg.IngredientsID)
	return err
}

const getAllIngredients = `-- name: GetAllIngredients :many
select id, name
from ingredients
order by id
`

func (q *Queries) GetAllIngredients(ctx context.Context) ([]Ingredient, error) {
	rows, err := q.db.Query(ctx, getAllIngredients)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Ingredient
	for rows.Next() {
		var i Ingredient
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIngredientNameByID = `-- name: GetIngredientNameByID :one
select name
from ingredients
where id = $1
`

func (q *Queries) GetIngredientNameByID(ctx context.Context, id int64) (string, error) {
	row := q.db.QueryRow(ctx, getIngredientNameByID, id)
	var name string
	err := row.Scan(&name)
	return name, err
}
