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

  primary_key {
    columns = [column.id]
  }

  index "tasks_name_idx" {
    columns = [column.name]
    unique  = false
  }
}