-- name: GetAllUnits :many
select id, name
from units
order by name;

-- name: AddUnit :one
insert into units (name)
values (sqlc.arg('name'))
on conflict (name) do update set name=excluded.name -- no-op that effectively does nothing, but returns the ID as intended
returning id;

-- name: DeleteUnit :exec
delete from units
where id = sqlc.arg('id');
