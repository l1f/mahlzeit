-- migrate:up
ALTER TABLE units ADD CONSTRAINT units_name_unique UNIQUE (name);

-- migrate:down
ALTER TABLE units DROP CONSTRAINT units_name_unique;
