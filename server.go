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
	"log"

	"github.com/graphql-go/graphql"
	//"github.com/adjust/rmq"
	"github.com/streadway/amqp"
)

type request struct {
	method string
	url   string
	body   string
	headers   string
	callback_url   string
	pid   string
	//executing bool
}

var tasks_queue []request

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
					fmt.Println("Came here :) req2")

					var req request = request{}

					if val, ok := p.Args["method"].(string); ok {
						req.method = val
					}

					if val, ok := p.Args["url"].(string); ok {
						req.url = val
					}

					if val, ok := p.Args["body"].(string); ok {
						req.body = val
					}

					if val, ok := p.Args["headers"].(string); ok {
						req.headers = val
					}

					if val, ok := p.Args["callback_url"].(string); ok {
						req.callback_url = val
					}

					req.pid = "SOME_RANDOM PID"
					


					var reque = append(tasks_queue, req)
					fmt.Println(tasks_queue)
					return reque, nil
					/*idQuery, isOK := p.Args["id"].(string)
					if isOK {
						return data[idQuery], nil
					}
					return nil, nil*/
				},
			},
		},
	})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

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

func main() {
	
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	
	q, err := ch.QueueDeclare(
	  "hello", // name
	  false,   // durable
	  false,   // delete when unused
	  false,   // exclusive
	  false,   // no-wait
	  nil,     // arguments
	)

	failOnError(err, "Failed to declare a queue")


	// BEGIN publishing to queue:
	body := "hello"
	err = ch.Publish(
	  "",     // exchange
	  q.Name, // routing key
	  false,  // mandatory
	  false,  // immediate
	  amqp.Publishing {
	    ContentType: "text/plain",
	    Body:        []byte(body),
	  })
	failOnError(err, "Failed to publish a message")
	// END publishing to queue.

	// BEGIN Consuming queue:
	msgs, err := ch.Consume(
	  q.Name, // queue
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
	    log.Printf("Received a message: %s", d.Body)
	  }
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	// END Consuming queue:

	//_ = importJSONDataFromFile("data.json", &data)

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Came here!!!")
		result := execute(r.URL.Query().Get("request"), schema)
		json.NewEncoder(w).Encode(result)
		fmt.Println("Finalizend")
		fmt.Println(tasks_queue)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got the result!")
	})

	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with Get      : curl -g 'http://localhost:8080/graphql?query={user(id:\"1\"){name}}'")
	http.ListenAndServe(":8080", nil)
}

func failOnError(err error, msg string) {
  if err != nil {
    fmt.Printf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}
/*
//Helper function to import json from file to map
func importJSONDataFromFile(fileName string, result interface{}) (isOK bool) {
	isOK = true
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Print("Error:", err)
		isOK = false
	}
	err = json.Unmarshal(content, result)
	if err != nil {
		isOK = false
		fmt.Print("Error:", err)
	}
	return
}
*/