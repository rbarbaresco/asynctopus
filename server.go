package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
)

type request struct {
	method string// `json:"name"`
	url   string// `json:"id"`
	body   string// `json:"id"`
	headers   string// `json:"id"`
	callback_url   string// `json:"id"`
	pid   string// `json:"id"`
}

var tasks_queue map[string]request

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
					fmt.Println("Came here :)")
					fmt.Println(p.Args)
					return "Some random pid.", nil
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
	//_ = importJSONDataFromFile("data.json", &data)

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Came here!!!")
		result := execute(r.URL.Query().Get("request"), schema)
		json.NewEncoder(w).Encode(result)
		fmt.Println("Finalizend")
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got the result!")
	})

	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with Get      : curl -g 'http://localhost:8080/graphql?query={user(id:\"1\"){name}}'")
	http.ListenAndServe(":8080", nil)
}

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