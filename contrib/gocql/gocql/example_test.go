package gocql_test

import (
	"context"

	gocqltrace "github.com/networknext/dd-trace-go/contrib/gocql/gocql"
	"github.com/networknext/dd-trace-go/tracer"
	"github.com/gocql/gocql"
)

// To trace Cassandra commands, use our query wrapper WrapQuery.
func Example() {
	// Initialise a Cassandra session as usual, create a query.
	cluster := gocql.NewCluster("127.0.0.1")
	session, _ := cluster.CreateSession()
	query := session.Query("CREATE KEYSPACE if not exists trace WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor': 1}")

	// Use context to pass information down the call chain
	root := tracer.NewRootSpan("parent.request", "web", "/home")
	ctx := root.Context(context.Background())

	// Wrap the query to trace it and pass the context for inheritance
	tracedQuery := gocqltrace.WrapQuery(query, gocqltrace.WithServiceName("ServiceName"))
	tracedQuery.WithContext(ctx)

	// Execute your query as usual
	tracedQuery.Exec()
}
