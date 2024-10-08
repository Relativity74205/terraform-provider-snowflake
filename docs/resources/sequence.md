---
page_title: "snowflake_sequence Resource - terraform-provider-snowflake"
subcategory: ""
description: |-
  
---

# snowflake_sequence (Resource)



## Example Usage

```terraform
resource "snowflake_database" "test" {
  name = "things"
}

resource "snowflake_schema" "test_schema" {
  name     = "things"
  database = snowflake_database.test.name
}

resource "snowflake_sequence" "test_sequence" {
  database = snowflake_database.test.name
  schema   = snowflake_schema.test_schema.name
  name     = "thing_counter"
}
```

-> **Note** Instead of using fully_qualified_name, you can reference objects managed outside Terraform by constructing a correct ID, consult [identifiers guide](https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest/docs/guides/identifiers#new-computed-fully-qualified-name-field-in-resources).
<!-- TODO(SNOW-1634854): include an example showing both methods-->

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) The database in which to create the sequence. Don't use the | character.
- `name` (String) Specifies the name for the sequence.
- `schema` (String) The schema in which to create the sequence. Don't use the | character.

### Optional

- `comment` (String) Specifies a comment for the sequence.
- `increment` (Number) The amount the sequence will increase by each time it is used
- `ordering` (String) The ordering of the sequence. Either ORDER or NOORDER. Default is ORDER.

### Read-Only

- `fully_qualified_name` (String) Fully qualified name of the resource. For more information, see [object name resolution](https://docs.snowflake.com/en/sql-reference/name-resolution).
- `id` (String) The ID of this resource.
- `next_value` (Number) The increment sequence interval.

## Import

Import is supported using the following syntax:

```shell
# format is database name | schema name | sequence name
terraform import snowflake_sequence.example 'dbName|schemaName|sequenceName'
```
