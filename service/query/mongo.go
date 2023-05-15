package query

/*
	Description:
		Package `query` provides interface for querying mongo db
		This pachage is basicly nothing but wrap https://github.com/mongodb/mongo-go-driver
		so please read document at following link for any detail
		https://godoc.org/go.mongodb.org/mongo-driver/mongo

	Use Case:
		Please Read the testcases for usage of each method
*/

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
)

var (
	// ErrNotFound is mongo document not found error
	ErrNotFound = fmt.Errorf("document not found")

	// ErrDuplicateKey is an error when violating unique index
	ErrDuplicateKey = fmt.Errorf("duplicate key")

	// ErrCollScan is error for unindexed query
	ErrCollScan = fmt.Errorf("COLLSCAN is not allowed")
)

type patchOp struct {
	patchMany bool
}

// PatchOp is an alias for functional argument
type PatchOp func(*patchOp)

// WithPatchMany specifies patchMany setting. To patch all entries selected, set patchMany = true.
func WithPatchMany(patchMany bool) PatchOp {
	return func(o *patchOp) {
		o.patchMany = patchMany
	}
}

type pipeOp struct {
	allowDiskUse bool
}

// PipeOp for options used in Pipe
type PipeOp func(*pipeOp)

// WithAllowDiskUse for `allowDiskUse`
// Enables writing to temporary files. document: https://docs.mongodb.com/manual/reference/command/aggregate/#cmd-aggregate-allowdiskuse
func WithAllowDiskUse(allow bool) PipeOp {
	return func(o *pipeOp) {
		o.allowDiskUse = true
	}
}

// UpsertOp is an upsert operation.
// TODO: make Upsert() use `UpsertOp` as the input parameter too
type UpsertOp struct {
	Selector interface{}
	Updater  interface{}
}

type CB func(context ctx.Ctx, raw bson.Raw, resumeToken bson.Raw) error

//Mongo abstract the mongo layer.
type Mongo interface {
	// Insert inserts a new document to the table
	Insert(context ctx.Ctx, table domain.Table, insert interface{}) error

	// FindOne get data from the table
	FindOne(context ctx.Ctx, table domain.Table, query, result interface{}) error

	// Count return counting for matched entry in the table
	// https://docs.mongodb.com/manual/reference/method/db.collection.countDocuments
	Count(context ctx.Ctx, table domain.Table, selector interface{}) (n int, err error)

	// CountEstimate return counting for matched entry in the table, it may not be accurate as Count but faster
	// https://docs.mongodb.com/manual/reference/method/db.collection.count/
	EstimateCount(context ctx.Ctx, table domain.Table, selector interface{}) (n int, err error)

	// Upsert update an entry , if the selector is already exist.
	// Upsert insert an entry , if the selector is not exist.
	Upsert(context ctx.Ctx, table domain.Table, selector, update interface{}) error

	// Search sort order by `sort` argument (ex "timestamp" ascending, or "-timestamp" descending)
	// if `sort` is "", the sort action is skipped, and the MongoDB does not guarantee the order of query results.
	Search(context ctx.Ctx, table domain.Table, offset, limit int, sort string, query, results interface{}) error

	// SearchNProject retrieves only the selected fields by `project` argument and sort by `sort` argument (ex "timestamp" ascending, or "-timestamp" descending)
	// if `sort` is "", the sort action is skipped, and the MongoDB does not guarantee the order of query results.
	SearchNProject(context ctx.Ctx, table domain.Table, offset, limit int, sort string, query, project, results interface{}) error

	// SearchNSorts sort with multiple fields, if you use compound key, make sure key order is correct. https://docs.mongodb.com/manual/tutorial/sort-results-with-indexes/
	SearchNSorts(context ctx.Ctx, table domain.Table, offset, limit int, sortFields []string, query, results interface{}) error

	// Remove remove an entry from the table
	// Return ErrNotFound if selector does not match any documents
	Remove(context ctx.Ctx, table domain.Table, selector interface{}) error

	//RemoveAll remove all entries matching the selector from the table
	RemoveAll(context ctx.Ctx, table domain.Table, selector interface{}) (removedCnt int64, err error)

	// Patch patch an entry, if the selector not exist, return err.
	// To patch all entries selected, set WithPatchMany(true).
	// Return ErrNotFound if selector does not match any documents
	Patch(context ctx.Ctx, table domain.Table, selector, update interface{}, ops ...PatchOp) error

	// CustomPatch patch an entry with customized mongo query
	// Return ErrNotFound if upsert is false and selector does not match any documents,
	CustomPatch(context ctx.Ctx, table domain.Table, selector, update bson.M, upsert bool) error

	// Increment let you increase a field number.
	// If entry not exist, insert it.
	Increment(context ctx.Ctx, table domain.Table, selector, result interface{}, field string, inc interface{}) error

	// IncrementMany let you increase fields and their values.
	// If entry not exist, insert with set statement.
	IncrementMany(context ctx.Ctx, table domain.Table, query interface{}, fieldAndValues bson.M, set bson.M, result interface{}) error

	// Push push `item` to `field` according `query`
	Push(context ctx.Ctx, table domain.Table, query, result interface{}, field string, item interface{}) error

	// Pull pull all `item` out from `field` according `query`
	Pull(context ctx.Ctx, table domain.Table, query, result interface{}, field string, item interface{}) error

	// Pipe wraps mongo's pipe function. It returns pipe as well as connection close
	// function. Caller needs to call fnClose to close session function after all iterations.
	Pipe(context ctx.Ctx, table domain.Table, pipeline interface{}, ops ...PipeOp) (p *Iter, fnClose func(), err error)

	// BulkUpsert performs multiple upsert operations.
	// Note that upsert operations are executed in parallel, as well as in a non-deterministic order.
	BulkUpsert(context ctx.Ctx, table domain.Table, BulkOps []UpsertOp) (matchedCnt int64, modifiedCnt int64, err error)

	RunWithTransaction(context ctx.Ctx, run func(ctx.Ctx) error) error
}
