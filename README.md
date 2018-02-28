# asynctopus

A docker application to turn synchronos requests into asynchronos calls. It's based on rabbitmq messaging system, so it's required you have a rabbitmq instance running on your machine. For development, I highly suggest you use a docker image for that. Check https://hub.docker.com/_/rabbitmq/.

## Main Technologies:
* GoLang
* GraphQL
* RabbitMQ
* Docker

Usage:

Option 1: docker-compose for local enviroment
1) build the application:  
$ docker-compose build

2) run the server:  
$ docker-compose up

3) access via browser http://localhost:8079/execute and you are good to go :)

The /execute route receives a parameter 'request' which is a GraphQL query with the given schema:  
```
request(  
  target_url: "http://example.com" // The url to make the service call  
  method: "GET" // The http method you wish to make the service call  
  body: "some body content" // If you are making a body based request, you may fill this field  
  headers: "Will be replaced by target_headers" // In the near future, it will be replaced by callback_headers.  
  callback_url: "http://myapplication.com/receiver" // The callback url which asynctopus must POST respond with the body content of the given request.  
  target_headers: "Not yet working" // In the near future, it will work, I promise!  
  callback_headers: "Not yet added" // In the near future, it will work, I promise!  
)
{pid} // It will tell asynctopus to return the pid for this request. Keep it safe so you can identify the response.
```
/execute?request={request(target_url:"http://example.com",method:"GET",callback_url:"http://myapplication.com/receiver"){pid}}

This request will return the pid for this call, which you will use to identify it later on your callback_url.
```json
{
  "data": {
    "request": {
      "pid": 1234
    }
  }
}
```

### IMPORTANT
Because asynctopus does not deal yet with the content-type of the response, it can't format it in the right way. So it will always respond to the callback_url with a byte response, as it's a generic way to send data.
