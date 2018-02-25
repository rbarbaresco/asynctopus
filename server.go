/*
https://hub.docker.com/_/redis/
docker run --name some-redis -d redis
*/
package main

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	//"time"
	//"log"
	"math/rand"

	"github.com/graphql-go/graphql"
	//"github.com/adjust/rmq"
	"github.com/streadway/amqp"
)

type Request struct {
	method string 			`json:"method"`
	url   string 			`json:"url"`
	body   string 			`json:"body"`
	headers   string 		`json:"headers"`
	callback_url   string 	`json:"callback_url"`
	pid   string 			`json:"pid"`
	//executing bool
}

var QUEUE_NAME = "asynctopus_task_queue"
var RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"
var CONSUMERS_SIZE = 5

/*
   Create User object type with fields "id" and "name" by using GraphQLObjectTypeConfig:
       - Name: name of object type
       - Fields: a map of fields by using GraphQLFields
   Setup type of field use GraphQLFieldConfig
*/
var requestType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Request",
		Fields: graphql.Fields{
			"method": &graphql.Field{
				Type: graphql.String,
			},
			"url": &graphql.Field{
				Type: graphql.String,
			},
			"body": &graphql.Field{
				Type: graphql.String,
			},
			"headers": &graphql.Field{
				Type: graphql.String,
			},
			"callback_url": &graphql.Field{
				Type: graphql.String,
			},
			"pid": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

/*
   Create Query object type with fields "user" has type [userType] by using GraphQLObjectTypeConfig:
       - Name: name of object type
       - Fields: a map of fields by using GraphQLFields
   Setup type of field use GraphQLFieldConfig to define:
       - Type: type of field
       - Args: arguments to query with current field
       - Resolve: function to query data using params from [Args] and return value with current type
*/
var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"request": &graphql.Field{
				Type: requestType,
				Args: graphql.FieldConfigArgument{
					"method": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"url": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"body": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"headers": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"callback_url": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"pid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					
					jsonObject, err := json.Marshal(createTask(p.Args))
					if err != nil {
						fmt.Println(err)
						return err, nil
					} else {
						fmt.Printf("json: %s\n", string(jsonObject))
						publish(jsonObject)
						return jsonObject, nil
					}
				},
			},
		},
	})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

func createTask(args map[string]interface {}) map[string]interface {} {
	var task = args
	task["pid"] = rand.Intn(1000)
	return task
}

func execute(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

func publish(message []byte) {

	conn, err := amqp.Dial(RABBITMQ_URL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	
	ch.QueueDeclare(
	  QUEUE_NAME, // name
	  false,   // durable
	  false,   // delete when unused
	  false,   // exclusive
	  false,   // no-wait
	  nil,     // arguments
	)

	// BEGIN publishing to queue:
	err = ch.Publish(
	  "",     // exchange
	  QUEUE_NAME, // routing key
	  false,  // mandatory
	  false,  // immediate
	  amqp.Publishing {
	    ContentType: "application/json",
	    Body:        message,
	  })
	failOnError(err, "Failed to publish a message")
	// END publishing to queue.
}

func startConsumers() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	
	ch.QueueDeclare(
	  QUEUE_NAME, // name
	  false,   // durable
	  false,   // delete when unused
	  false,   // exclusive
	  false,   // no-wait
	  nil,     // arguments
	)

	// BEGIN Consuming queue:
	msgs, err := ch.Consume(
	  QUEUE_NAME, // queue
	  "",     // consumer
	  true,   // auto-ack
	  false,  // exclusive
	  false,  // no-local
	  false,  // no-wait
	  nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
	  for d := range msgs {
	    consume(d.Body)
	  }
	}()

	fmt.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	// END Consuming queue:
}

func consume(message []byte) {
    fmt.Printf("Received a message: %s", message)
	/*client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
	}

	resp, err := client.Get("http://example.com")
	// ...

	req, err := http.NewRequest("GET", "http://example.com", nil)
	// ...
	req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := client.Do(req)
	*/
}

func main() {

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		result := execute(r.URL.Query().Get("request"), schema)
		json.NewEncoder(w).Encode(result)

	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got the result!")
	})
	
	go startConsumers()

	fmt.Println("Now server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func failOnError(err error, msg string) {
  if err != nil {
    fmt.Printf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}
