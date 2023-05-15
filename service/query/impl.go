package query

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/domain"
)

const (
	tableAttribute = "table"
	queryMaxTime   = 20 * time.Second
	maxQueueSize   = 30
)

var (
	timeNow = time.Now
)

type impl struct {
	client     *mongoclient.Client
	checkIndex bool
	tokens     chan int
}

// New initializes an impl
func New(client *mongoclient.Client, checkIndex bool) Mongo {
	limit := 10
	tokens := make(chan int, limit)
	for i := 0; i < limit; i++ {
		tokens <- i + 1
	}
	return &impl{
		client:     client,
		checkIndex: checkIndex,
		tokens:     tokens,
	}
}

func (im *impl) logerr(context ctx.Ctx, msg string, err error) {
	if _, ok := err.(topology.ConnectionError); ok {
		// met.BumpSum("conn.err", 1.0)
	}
	context.WithFields(log.Fields{"err": err}).Error(msg)

}

func (im *impl) getClient(context ctx.Ctx) *mongoclient.Client {
	return im.client
}

func (im *impl) Insert(context ctx.Ctx, table domain.Table, insert interface{}) error {
	// defer met.BumpTime("time", "func", "insert", "table", string(table)).End()
	defer slowLog(context, string(table), "insert", nil, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":  table,
		"insert": insert,
	})

	if _, err := client.Database(client.DbName).Collection(string(table)).InsertOne(context, insert); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		im.logerr(context, "Insert: InsertOne failed", err)
		return err
	}

	return nil
}

func (im *impl) FindOne(context ctx.Ctx, table domain.Table, query, result interface{}) error {
	// defer met.BumpTime("time", "func", "findone", "table", string(table)).End()
	defer slowLog(context, string(table), "findone", query, nil)()
	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table": table,
		"query": query,
	})

	if err := im.checkQueryIndex(context, string(table), "find", bson.E{Key: "filter", Value: query}); err != nil {
		im.logerr(context, "checkQueryIndex failed", err)
		return err
	}

	findOneOpts := options.FindOne().SetMaxTime(queryMaxTime)
	res := client.Database(client.DbName).Collection(string(table)).FindOne(context, query, findOneOpts)

	if err := res.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		im.logerr(context, "FindOne: FindOne error", err)
		return err
	}
	return nil
}

func (im *impl) Count(context ctx.Ctx, table domain.Table, selector interface{}) (n int, err error) {
	// defer met.BumpTime("time", "func", "count", "table", string(table)).End()
	defer slowLog(context, string(table), "count", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
	})

	if err := im.checkQueryIndex(context, string(table), "count", bson.E{Key: "query", Value: selector}); err != nil {
		im.logerr(context, "checkQueryIndex failed", err)
		return 0, err
	}

	opts := options.Count().SetMaxTime(queryMaxTime)
	count, err := client.Database(client.DbName).Collection(string(table)).CountDocuments(context, selector, opts)
	if err != nil {
		im.logerr(context, "Count: CountDocuments failed", err)
		return 0, err
	}
	return int(count), nil
}

func (im *impl) EstimateCount(context ctx.Ctx, table domain.Table, selector interface{}) (n int, err error) {
	// defer met.BumpTime("time", "func", "estimateCount", "table", string(table)).End()
	defer slowLog(context, string(table), "estimateCount", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
	})

	if err := im.checkQueryIndex(context, string(table), "count", bson.E{Key: "query", Value: selector}); err != nil {
		im.logerr(context, "checkQueryIndex failed", err)
		return 0, err
	}

	cmd := bson.D{
		bson.E{Key: "count", Value: table},
		bson.E{Key: "query", Value: selector},
	}
	res, err := client.Database(client.DbName).RunCommand(context, cmd).DecodeBytes()
	if err != nil {
		im.logerr(context, "RunCommand failed", err)
		return 0, err
	}
	count := res.Lookup("n").Int32()
	return int(count), nil
}

func (im *impl) Upsert(context ctx.Ctx, table domain.Table, selector, update interface{}) error {
	// defer met.BumpTime("time", "func", "upsert", "table", string(table)).End()
	defer slowLog(context, string(table), "upsert", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
		"update":   update,
	})

	replaceOpts := options.Replace().SetUpsert(true)
	if _, err := client.Database(client.DbName).Collection(string(table)).ReplaceOne(context, selector, update, replaceOpts); err != nil {
		im.logerr(context, "Upsert: ReplaceOne failed", err)
		return err
	}
	return nil
}

func (im *impl) getSortOption(context ctx.Ctx, sortStrings ...string) bson.D {
	res := bson.D{}
	for _, sort := range sortStrings {
		if sort == "" {
			continue
		}
		if sort[0] == '-' {
			res = append(res, bson.E{Key: sort[1:], Value: -1})
		} else {
			res = append(res, bson.E{Key: sort, Value: 1})
		}
	}

	return res
}

func (im *impl) search(context ctx.Ctx, table domain.Table, offset, limit int, sortFields []string, query, project, results interface{}) error {
	// defer met.BumpTime("time", "func", "search", "table", string(table)).End()
	defer slowLog(context, string(table), "search", query, sortFields)()
	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table": table,
		"query": query,
	})

	if err := im.checkQueryIndex(context, string(table), "find", bson.E{Key: "filter", Value: query}); err != nil {
		im.logerr(context, "checkQueryIndex failed", err)
		return err
	}

	findOpts := options.Find().SetMaxTime(queryMaxTime)
	findOpts.SetLimit(int64(limit)).SetSkip(int64(offset))
	sortOpt := im.getSortOption(context, sortFields...)
	if len(sortOpt) > 0 {
		findOpts.SetSort(sortOpt)
	}
	if project != nil {
		findOpts.SetProjection(project)
	}
	cursor, err := client.Database(client.DbName).Collection(string(table)).Find(context, query, findOpts)
	if err != nil {
		im.logerr(context, "Search: Find failed", err)
		return err
	}
	defer cursor.Close(context)

	if err := cursor.All(context, results); err != nil {
		im.logerr(context, "Search: cursor.All failed", err)
		return err
	}
	return nil
}

func (im *impl) Search(context ctx.Ctx, table domain.Table, offset, limit int, sort string, query, results interface{}) error {
	return im.SearchNProject(context, table, offset, limit, sort, query, nil, results)
}

func (im *impl) SearchNProject(context ctx.Ctx, table domain.Table, offset, limit int, sort string, query, project, results interface{}) error {
	return im.search(context, table, offset, limit, []string{sort}, query, project, results)
}

func (im *impl) SearchNSorts(context ctx.Ctx, table domain.Table, offset, limit int, sortFields []string, query, results interface{}) error {
	return im.search(context, table, offset, limit, sortFields, query, nil, results)
}

func (im *impl) Remove(context ctx.Ctx, table domain.Table, selector interface{}) error {
	// defer met.BumpTime("time", "func", "remove", "table", string(table)).End()
	defer slowLog(context, string(table), "remove", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
	})

	if deletedRes, err := client.Database(client.DbName).Collection(string(table)).DeleteOne(context, selector); err != nil {
		im.logerr(context, "Remove: DeleteOne failed", err)
		return err
	} else if deletedRes.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (im *impl) RemoveAll(context ctx.Ctx, table domain.Table, selector interface{}) (int64, error) {
	// defer met.BumpTime("time", "func", "removeAll", "table", string(table)).End()
	defer slowLog(context, string(table), "removeAll", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
	})

	res, err := client.Database(client.DbName).Collection(string(table)).DeleteMany(context, selector)
	if err != nil {
		im.logerr(context, "RemoveAll: DeleteMany failed", err)
		return 0, err
	}

	return res.DeletedCount, nil
}

func initPatchOp() *patchOp {
	return &patchOp{}
}

func (im *impl) Patch(context ctx.Ctx, table domain.Table, selector, update interface{}, ops ...PatchOp) error {
	// defer met.BumpTime("time", "func", "update", "table", string(table)).End()
	defer slowLog(context, string(table), "update", selector, nil)()

	// load options
	o := initPatchOp()
	for _, opt := range ops {
		opt(o)
	}

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
		"update":   update,
	})

	var err error
	var updateRes *mongo.UpdateResult
	updater := bson.M{"$set": update}
	if o.patchMany {
		updateRes, err = client.Database(client.DbName).Collection(string(table)).UpdateMany(context, selector, updater)
		if err != nil {
			im.logerr(context, "Patch: UpdateMany failed", err)
			return err
		}
	} else {
		updateRes, err = client.Database(client.DbName).Collection(string(table)).UpdateOne(context, selector, updater)
		if err != nil {
			im.logerr(context, "Patch: UpdateOne failed", err)
			return err
		}
	}

	if updateRes.MatchedCount == 0 && updateRes.ModifiedCount == 0 && updateRes.UpsertedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (im *impl) CustomPatch(context ctx.Ctx, table domain.Table, selector, update bson.M, upsert bool) error {
	// defer met.BumpTime("time", "func", "customupdate", "table", string(table)).End()
	defer slowLog(context, string(table), "customupdate", selector, nil)()

	client := im.getClient(context)

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"selector": selector,
		"update":   update,
	})

	updateOpts := options.Update().SetUpsert(upsert)
	updateRes, err := client.Database(client.DbName).Collection(string(table)).UpdateOne(context, selector, update, updateOpts)
	if err != nil {
		im.logerr(context, "CustomPatch: UpdateOne failed", err)
		return err
	}

	if updateRes.MatchedCount == 0 && updateRes.ModifiedCount == 0 && updateRes.UpsertedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (im *impl) Increment(context ctx.Ctx, table domain.Table, selector, result interface{}, field string, inc interface{}) error {
	return im.IncrementMany(context, table, selector, bson.M{field: inc}, nil, result)
}

func (im *impl) IncrementMany(context ctx.Ctx, table domain.Table, query interface{}, fieldAndValues bson.M, set bson.M, result interface{}) error {
	// defer met.BumpTime("time", "func", "incrementMany", "table", string(table)).End()
	defer slowLog(context, string(table), "incrementMany", query, nil)()

	client := im.getClient(context)

	updater := bson.M{"$inc": fieldAndValues}
	if set != nil {
		updater["$setOnInsert"] = set
	}
	findOneAndUpdateOpts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	res := client.Database(client.DbName).Collection(string(table)).FindOneAndUpdate(context, query, updater, findOneAndUpdateOpts)
	if err := res.Decode(result); err != nil {
		im.logerr(context, "IncrementMany: FindOneAndUpdate failed", err)
		return err
	}
	return nil
}

func (im *impl) Push(context ctx.Ctx, table domain.Table, query, result interface{}, field string, item interface{}) error {
	// defer met.BumpTime("time", "func", "push", "table", string(table)).End()
	defer slowLog(context, string(table), "push", query, nil)()

	client := im.getClient(context)

	updater := bson.M{"$push": bson.M{field: item}}
	findOneAndUpdateOpts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	res := client.Database(client.DbName).Collection(string(table)).FindOneAndUpdate(context, query, updater, findOneAndUpdateOpts)
	if err := res.Decode(result); err != nil {
		im.logerr(context, "Push: FindOneAndUpdate failed", err)
		return err
	}
	return nil
}

func (im *impl) Pull(context ctx.Ctx, table domain.Table, query, result interface{}, field string, item interface{}) error {
	// defer met.BumpTime("time", "func", "pull", "table", string(table)).End()
	defer slowLog(context, string(table), "pull", query, nil)()

	client := im.getClient(context)

	updater := map[string]interface{}{"$pull": map[string]interface{}{field: item}}
	findOneAndUpdateOpts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	res := client.Database(client.DbName).Collection(string(table)).FindOneAndUpdate(context, query, updater, findOneAndUpdateOpts)
	if err := res.Decode(result); err != nil {
		im.logerr(context, "Pull: FindOneAndUpdate failed", err)
		return err
	}
	return nil
}

func (im *impl) BulkUpsert(context ctx.Ctx, table domain.Table, upsertOps []UpsertOp) (matchedCnt int64, modifiedCnt int64, err error) {
	// defer met.BumpTime("time", "func", "bulkUpsert", "table", string(table)).End()

	if len(upsertOps) == 0 {
		return 0, 0, fmt.Errorf("length of `pairs` equals 0")
	}

	client := im.getClient(context)
	bulkWriteOpts := options.BulkWrite().SetOrdered(false)
	models := make([]mongo.WriteModel, 0, len(upsertOps))
	for _, op := range upsertOps {
		models = append(models, mongo.NewReplaceOneModel().SetFilter(op.Selector).SetReplacement(op.Updater).SetUpsert(true))
	}
	res, err := client.Database(client.DbName).Collection(string(table)).BulkWrite(context, models, bulkWriteOpts)
	if err != nil {
		im.logerr(context, "BulkUpsert: BulkWrite failed", err)
		return 0, 0, err
	}
	return res.MatchedCount, res.ModifiedCount, nil
}

// Iter wraps mongo's cursor struct
type Iter struct {
	cursor *mongo.Cursor
	table  domain.Table
}

// Next overrides mongo's iter next method with default value insertion
func (iter *Iter) Next(context ctx.Ctx, result interface{}) (bool, error) {
	// defer met.BumpTime("time", "func", "next", "table", string(iter.table)).End()
	ok := iter.cursor.Next(context)
	if !ok {
		return false, nil
	}
	if err := iter.cursor.Decode(result); err != nil {
		return false, err
	}
	return true, nil
}

func (iter *Iter) All(context ctx.Ctx, result interface{}) error {
	// defer met.BumpTime("time", "func", "all", "table", string(iter.table)).End()
	if err := iter.cursor.All(context, result); err != nil {
		return err
	}
	return nil
}

func initPipeOp() *pipeOp {
	return &pipeOp{}
}

func (im *impl) Pipe(context ctx.Ctx, table domain.Table, pipeline interface{}, ops ...PipeOp) (*Iter, func(), error) {
	// defer met.BumpTime("time", "func", "pipe", "table", string(table)).End()
	defer slowLog(context, string(table), "pipe", pipeline, nil)()

	o := initPipeOp()
	for _, op := range ops {
		op(o)
	}

	client := im.getClient(context)

	// TODO: check pipe index

	context = ctx.WithValues(context, map[string]interface{}{
		"table":    table,
		"pipeline": pipeline,
	})

	aggregateOpts := options.Aggregate().SetAllowDiskUse(o.allowDiskUse)

	cursor, err := client.Database(client.DbName).Collection(string(table)).Aggregate(context, pipeline, aggregateOpts)
	if err != nil {
		im.logerr(context, "Pipe: Aggregate failed", err)
		return nil, nil, err
	}

	return &Iter{cursor: cursor, table: table},
		func() { cursor.Close(ctx.Background()) }, nil
}

func (im *impl) RunWithTransaction(context ctx.Ctx, run func(ctx.Ctx) error) error {
	var token int
	select {
	case <-context.Done():
	case token = <-im.tokens:
	}
	defer func() {
		if token != 0 {
			im.tokens <- token
		}
	}()

	client := im.getClient(context)

	// explain command is not support in transaction
	if im.checkIndex {
		return run(context)
	}

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context)

	fn := func(sessCtx mongo.SessionContext) (interface{}, error) {
		c := ctx.Ctx{
			Context: sessCtx,
			Logger:  context.Logger,
		}
		return nil, run(c)
	}
	_, err = session.WithTransaction(context, fn)
	return err
}

func slowLog(context ctx.Ctx, table, action string, query interface{}, sort interface{}) func() {
	start := timeNow()
	threshold := int64(500)

	return func() {
		elapsed := time.Since(start)
		elapsedMs := elapsed.Nanoseconds() / time.Millisecond.Nanoseconds()
		if elapsedMs >= threshold {
			// // met.BumpSum("mongo.slowlog", 1, "table", table, "action", action)
			context.WithFields(log.Fields{
				"table":        table,
				"action":       action,
				"startTimeStr": start,
				"startTime":    start.Unix(),
				"durationMs":   elapsedMs,
				"query":        query,
				"sort":         sort,
			}).Warn("mongo slowlog")
		}
	}
}

func (im *impl) checkQueryIndex(context ctx.Ctx, table string, action string, query bson.E) error {
	if !im.checkIndex {
		return nil
	}
	// reference: https://docs.mongodb.com/manual/reference/command/explain/
	client := im.getClient(context)
	res := client.Database(client.DbName).RunCommand(context, bson.D{
		bson.E{
			Key: "explain",
			Value: bson.D{
				bson.E{Key: action, Value: table},
				query,
			},
		},
		bson.E{
			Key:   "verbosity",
			Value: "queryPlanner",
		},
	})

	var m bson.M
	if err := res.Decode(&m); err != nil {
		context.WithField("err", err).Warn("checkQueryIndex decode failed")
		// met.BumpSum("checkQueryIndex.err", float64(1))
		return nil
	}

	// We only check if `COLLSCAN` is in `m` as string since the data structure
	// of `m` is not consistent for all environment. It's quite difficult to use
	// struct to marshal `m`.
	if strings.Contains(fmt.Sprintf("%v", m), "COLLSCAN") {
		context.WithField("query", query).Warn("COLLSCAN")
		return ErrCollScan
	}
	return nil
}
