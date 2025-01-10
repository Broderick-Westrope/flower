schema "main" {}

table "tasks" {
  schema = schema.main
  column "id" {
    type = integer
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "description" {
    type = text
    null = false
  }
  column "parent_id" {
    type = integer
    null = true
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_tasks_parent" {
    columns     = [column.parent_id]
    ref_columns = [column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }
}

table "sessions" {
  schema = schema.main
  column "id" {
    type = integer
    null = false
  }
  column "task_id" {
    type = integer
    null = false
  }
  column "started_at" {
    type = integer
    null = false
  }
  column "ended_at" {
    type = integer
    null = false
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_tasks_sessions" {
    columns     = [column.task_id]
    ref_columns = [table.tasks.column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }
}