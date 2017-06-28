package sql_datastore

// Cmd represents a set of standardized SQL queries these abstractions
// define a common set of commands that a model can provide to sql_datastore
// for execution
type Cmd int

const (
	// Unknown as default, errored state
	CmdUnknown Cmd = iota
	// starting with DDL statements:
	// CREATE TABLE query
	CmdCreateTable
	// ALTER TABLE statement
	CmdAlterTable
	// DROP TABLE statement
	CmdDropTable
	// SELECT statement that should return a single result
	CmdSelectOne
	// INSERT a single row
	CmdInsertOne
	// UPDATE a single row
	CmdUpdateOne
	// DELETE a single row
	CmdDeleteOne
	// Check if a single row exists
	CmdExistsOne
	// List Query, can return many rows. List is special
	// in that LIMIT & OFFSET params are provided by the query
	// method. See note in Datastore.Query()
	CmdList
)
