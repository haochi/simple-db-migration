# simple-db-migration

The script will create a `simple_db_migration` table in your database to keep track of the applied migration files.

## Examples

```bash
# Run all new migrations (to execute all the unapplied "*.up.sql")
./simple-db-migration -config config.json
```

```
# Undo a migration (to execute the file "2.*.down.sql")
./simple-db-migration -config config.json -down 2
```