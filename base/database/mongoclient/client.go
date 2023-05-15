package mongoclient

import (
	"context"
	"crypto/tls"
	"runtime"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/x-xyz/goapi/base/log"
)

const (
	mgSocketTimeout = 60 * time.Second
	mgoPoolLimit    = 4096
)

// Client wraps mongo.Client
type Client struct {
	DbName string
	*mongo.Client
}

// MustConnectMongoClient returns MongoDB connection client if connected successfully, or it will trigger panic
func MustConnectMongoClient(uri, authDBName, dbName string, ssl, setSafe bool, poolSizeMultiplier float64) *Client {
	cli, err := ConnectMongoClient(uri, authDBName, dbName, ssl, setSafe, poolSizeMultiplier)
	if err != nil {
		log.Log().WithFields(log.Fields{"mongoURI": uri, "err": err}).Panic("fail to dial Mongo")
	}
	return cli
}

// ConnectMongoClient returns mongo driver client
func ConnectMongoClient(uri, authDBName, dbName string, ssl, setSafe bool, poolSizeMultiplier float64) (*Client, error) {

	ctx := context.Background()
	connSetting, err := connstring.Parse(uri)
	if err != nil {
		log.Log().WithFields(log.Fields{
			"mongoURI": uri,
			"dbName":   dbName,
			"err":      err,
		}).Error("fail to parse connstring")
		return nil, err
	}

	clientOpts := options.Client()
	clientOpts.ApplyURI(uri)
	clientOpts.SetSocketTimeout(mgSocketTimeout)

	// If AuthSource is not set in connstring, set it to authDBName
	if connSetting.Username != "" && connSetting.AuthSource == "" {
		clientOpts.SetAuth(options.Credential{
			AuthMechanism:           connSetting.AuthMechanism,
			AuthMechanismProperties: connSetting.AuthMechanismProperties,
			Username:                connSetting.Username,
			Password:                connSetting.Password,
			PasswordSet:             connSetting.PasswordSet,
			AuthSource:              authDBName,
		})
	}

	// total connection pool size
	poolSize := int(float64(runtime.NumCPU()) * poolSizeMultiplier)
	// because each host has its own connection pool,
	// if we set poolSize directly, it generate too many connections,
	// we set each host's pool size by divide the number of hosts
	poolSize = (poolSize + len(connSetting.Hosts) - 1) / len(connSetting.Hosts)
	clientOpts.SetMinPoolSize(uint64(poolSize / 4))
	clientOpts.SetMaxPoolSize(uint64(poolSize))
	log.Log().WithField("poolSize", poolSize).Info("mongo driver pool size")

	// FIXME: too many logs, using metrics instead
	// poolMonitor := event.PoolMonitor{Event: func(evt *event.PoolEvent) {
	// 	log.Log().WithFields(log.Fields{
	// 		"eventType":   evt.Type,
	// 		"eventResion": evt.Reason,
	// 	}).Info("mongo pool event")
	// }}
	// clientOpts.SetPoolMonitor(&poolMonitor)

	if ssl {
		tlsConfig := &tls.Config{}
		clientOpts.SetTLSConfig(tlsConfig)
	}

	if setSafe {
		// Force the server to wait for a majority of members of a replica set to return
		clientOpts.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	}
	clientOpts.SetRetryWrites(true)

	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		log.Log().WithFields(log.Fields{
			"mongoHosts": connSetting.Hosts,
			"dbName":     dbName,
			"err":        err,
		}).Error("fail to create mongo client")
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		log.Log().WithFields(log.Fields{
			"mongoHosts": connSetting.Hosts,
			"dbName":     dbName,
			"err":        err,
		}).Error("fail to connect mongo db")
		return nil, err
	}

	// Test if mongoDBName is valid
	if _, err := client.Database(dbName).ListCollectionNames(ctx, bson.D{}); err != nil {
		log.Log().WithFields(log.Fields{
			"mongoHosts": connSetting.Hosts,
			"dbName":     dbName,
			"err":        err,
		}).Error("fail to test mongo db")
		return nil, err
	}

	log.Log().WithFields(log.Fields{
		"mongoHosts": connSetting.Hosts,
		"db":         dbName,
	}).Info("mongo connected")
	return &Client{
		Client: client,
		DbName: dbName,
	}, nil
}
