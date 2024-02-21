# Getting Started with Camunda Cloud using Go
Getting started with Camunda using Golang and Docker

## Create Camunda Cloud cluster
* Log in to [https://console.cloud.camunda.io](https://console.cloud.camunda.io).
* Choose Cluster from header menu and create a new Cluster
![process-image](https://github.com/vietnguyendl61/go-camunda-getting-started/blob/main/resources/images/header_menu.png)
* When the new cluster appears in the console, create a new set of client credentials in tab API.
![create-client-credentials](https://github.com/vietnguyendl61/go-camunda-getting-started/blob/main/resources/images/create_client_credentials.png)
* Copy the client Connection Info environment variables block.

## Configure connection

We will use [GoDotEnv](https://github.com/joho/godotenv) to environmental the client connection credentials.

* Add GoDotEnv to the project:

```bash
go get github.com/joho/godotenv
```

* Add the client connection credentials for your cluster to the file `.env`:

**Note**: _make sure to remove the `export` keyword from each line_.

```
ZEEBE_ADDRESS='a6c3e854-376f-456b-b7d2-508291ed5f05.syd-1.zeebe.camunda.io:443'
ZEEBE_CLIENT_ID='4TST-DF5Myx~TASVos-iIDzoz12-xsE5'
ZEEBE_CLIENT_SECRET='tSiNmwAcSJmfEESztf7Ni~6pYm8Dr-z-DQGRS4-0zhKW-GzHPPuT.Gg_DkDmYIWF'
ZEEBE_AUTHORIZATION_SERVER_URL='https://login.cloud.camunda.io/oauth/token'
ZEEBE_TOKEN_AUDIENCE='zeebe.camunda.io'
CAMUNDA_CLUSTER_ID='a6c3e854-376f-456b-b7d2-508291ed5f05'
CAMUNDA_CLUSTER_REGION='syd-1'
CAMUNDA_CREDENTIALS_SCOPES='Zeebe,Tasklist,Operate,Optimize'
CAMUNDA_TASKLIST_BASE_URL='https://syd-1.tasklist.camunda.io/a6c3e854-376f-456b-b7d2-508291ed5f05'
CAMUNDA_OPTIMIZE_BASE_URL='https://syd-1.optimize.camunda.io/a6c3e854-376f-456b-b7d2-508291ed5f05'
CAMUNDA_OPERATE_BASE_URL='https://syd-1.operate.camunda.io/a6c3e854-376f-456b-b7d2-508291ed5f05'
CAMUNDA_OAUTH_URL='https://login.cloud.camunda.io/oauth/token'
```

* Save the file.

## Test Connection with Camunda Cloud

* Paste the following code into the file `main.go`:

```go
package main

import (
	"context"
	"github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Error when load env: " + err.Error())
        return
    }
    client, err := zbc.NewClient(&zbc.ClientConfig{
        GatewayAddress: os.Getenv("ZEEBE_ADDRESS"),
    })
    
    if err != nil {
        log.Println("Error when create client: " + err.Error())
        return
    }
    
    ctx := context.Background()
    err = CheckConnection(ctx, client)
    if err != nil {
        log.Println("Error when check connection: " + err.Error())
        return
    }
}

func CheckConnection(ctx context.Context, client zbc.Client) error {
    topology, err := client.NewTopologyCommand().Send(ctx)
    if err != nil {
        return err
    }
    
    for _, broker := range topology.Brokers {
        log.Println("Broker", broker.Host, ":", broker.Port)
        for _, partition := range broker.Partitions {
            log.Println("  Partition", partition.PartitionId, ":", roleToString(partition.Role))
        }
    }
    return nil
}

func roleToString(role pb.Partition_PartitionBrokerRole) string {
    switch role {
    case pb.Partition_LEADER:
        return "Leader"
    case pb.Partition_FOLLOWER:
        return "Follower"
    default:
        return "Unknown"
    }
}
```

* Run the program with the command `go run main.go`.

* You will see output similar to the following:

```
2024/02/20 14:34:06 Broker zeebe-2.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/20 14:34:06   Partition 1 : Follower
2024/02/20 14:34:06   Partition 2 : Follower
2024/02/20 14:34:06   Partition 3 : Leader
2024/02/20 14:34:06 Broker zeebe-1.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/20 14:34:06   Partition 1 : Follower
2024/02/20 14:34:06   Partition 2 : Leader
2024/02/20 14:34:06   Partition 3 : Follower
2024/02/20 14:34:06 Broker zeebe-0.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/20 14:34:06   Partition 1 : Leader
2024/02/20 14:34:06   Partition 2 : Follower
2024/02/20 14:34:06   Partition 3 : Follower
```

This is the topology response from the cluster.

## Create a BPMN model

* Download and install the [Camunda Modeler](https://camunda.com/download/modeler).
* Open Camunda Modeler and create a new BPMN Diagram.
* Create a new BPMN diagram. Flow the step below:
  *  Add a few service tasks to the BPMN diagram and set the required attributes.
  *  These are created when the process instance reaches a service task.
  *  Open the BPMN diagram in Modeler. Keeping in mind how you want to deploy your model, you can choose either Web Modeler or Desktop Modeler
  *  Insert three service tasks between the start and the end event.
     * Name the first task `Collect Money`.
     * Name the second task `Fetch Items`.
     * Name the third task `Ship Parcel`.
  *  Using the properties panel Task definition section, set the type of each task, which identifies the nature of the work to be performed.
     * Set the **type** of the first task to `payment-service`.
     * Set the **type** of the second task to `fetcher-service`.
     * Set the **type** of the third task to `shipping-service`.
  *  Additionally, for the service task `Collect Money` set a **task-header** with the key method and the value `VISA`. This header is used as a configuration parameter for the payment-service worker to hand over the payment method.

It should look like this:

![process-collect-money](https://github.com/vietnguyendl61/go-camunda-getting-started/blob/main/resources/images/process.png)

* Save the diagram to `order-process.bpmn` in your project.

## Deploy the BPMN model to Camunda Cloud

* Edit the `main.go` file, and add a new function `deploy`:

```go
func DeployResource(ctx context.Context, client zbc.Client) error {
    response, err := client.NewDeployResourceCommand().
		AddResourceFile("resources/order-process.bpmn").Send(ctx)
    if err != nil {
        return err
    }
    log.Println(response.String())
    return nil
}
```

* Now update the `main()` function to look like this:

```go
func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Error when load env: " + err.Error())
        return
    }
    client, err := zbc.NewClient(&zbc.ClientConfig{
        GatewayAddress: os.Getenv("ZEEBE_ADDRESS"),
    })
    
    if err != nil {
        log.Println("Error when create client: " + err.Error())
        return
    }
    
    ctx := context.Background()
    err = CheckConnection(ctx, client)
    if err != nil {
        log.Println("Error when check connection: " + err.Error())
        return
    }
    
    err = DeployResource(ctx, client)
    if err != nil {
        log.Println("Error when deploy resource: " + err.Error())
        return
    }
}
```
* Run the program with `go run main.go`.

You will see the deployment response:

```
2024/02/21 09:13:04 Broker zeebe-2.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/21 09:13:04   Partition 1 : Follower
2024/02/21 09:13:04   Partition 2 : Follower
2024/02/21 09:13:04   Partition 3 : Leader
2024/02/21 09:13:04 Broker zeebe-1.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/21 09:13:04   Partition 1 : Follower
2024/02/21 09:13:04   Partition 2 : Leader
2024/02/21 09:13:04   Partition 3 : Follower
2024/02/21 09:13:04 Broker zeebe-0.zeebe-broker-service.a6c3e854-376f-456b-b7d2-508291ed5f05-zeebe.svc.cluster.local : 26501
2024/02/21 09:13:04   Partition 1 : Leader
2024/02/21 09:13:04   Partition 2 : Follower
2024/02/21 09:13:04   Partition 3 : Follower
2024/02/21 09:13:05 key:2251799813831538 deployments:{process:{bpmnProcessId:"Process_1vepm8y" version:2 processDefinitionKey:2251799813831537 resourceName:"resources/order-process.bpmn" 5:"<default>"}} 3:"<default>" 
```

## Create a process instance

You are ready to create our first instance of the deployed process.

A process instance is created by a specific version of the process, which can be set on creation.

* Edit file `main.go`, add the function below:
```go
func CreateProcessInstance(ctx context.Context, client zbc.Client) error {
    variables := make(map[string]interface{})
    variables["orderId"] = "31243"
    
    request, err := client.NewCreateInstanceCommand().BPMNProcessId("Process_1vepm8y").LatestVersion().VariablesFromMap(variables)
    if err != nil {
        return err
    }
    
    msg, err := request.Send(ctx)
    if err != nil {
		return err
    }
    
    log.Println(msg.String())
    return nil
}
```
* Update the `main()` function:
```go
func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Error when load env: " + err.Error())
        return
    }
    client, err := zbc.NewClient(&zbc.ClientConfig{
        GatewayAddress: os.Getenv("ZEEBE_ADDRESS"),
    })
    
    if err != nil {
        log.Println("Error when create client: " + err.Error())
        return
    }
    
    ctx := context.Background()
    err = CheckConnection(ctx, client)
    if err != nil {
        log.Println("Error when check connection: " + err.Error())
        return
    }
    
    err = DeployResource(ctx, client)
    if err != nil {
        log.Println("Error when deploy resource: " + err.Error())
        return
    }
    
    err = CreateProcessInstance(ctx, client)
    if err != nil {
        log.Println("Error when create process instance: " + err.Error())
        return
    }
}
```

Run the program and verify the process instance is created. You will note an output similar to below:
```
2024/02/21 09:22:54 processDefinitionKey:2251799813831537  bpmnProcessId:"Process_1vepm8y"  version:2  processInstanceKey:4503599627517021  5:"<default>"
```

## View a Workflow Instance in Operate

Want to note how the process instance is executed? Follow the steps below:

* Go to the cluster in Camunda 8 and select it.
* Click on the link to [Operate](https://docs.camunda.io/docs/components/operate/userguide/basic-operate-navigation/).
* Select the process order process.

Let's create a task worker to serve the job represented by this task.

## Create a Job Worker

First, we open a [job worker](https://docs.camunda.io/docs/components/concepts/job-workers/) for jobs of type `payment-service`.

The job worker will repeatedly poll for new jobs of the type payment-service and activate them subsequently. Each activated job will then be passed to the job handler, which implements the business logic of the job worker.

The handler will then complete the job with its result or fail the job if it encounters a problem while processing the job.

When observing the current state of the process in Operate, you can note the process instance moved from the first service task to the next one.

* Edit the `main.go` file, add 2 functions:
```go
func handleJobPaymentService(client worker.JobClient, job entities.Job) {
    jobKey := job.GetKey()
    
    headers, err := job.GetCustomHeadersAsMap()
    if err != nil {
        // failed to handle job as we require the custom job headers
        failJob(client, job)
        return
    }
    
    variables, err := job.GetVariablesAsMap()
    if err != nil {
        // failed to handle job as we require the variables
        failJob(client, job)
        return
    }
    
    variables["totalPrice"] = 46.50
    request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
    if err != nil {
        // failed to set the updated variables
        failJob(client, job)
        return
    }
    
    log.Println("Complete job", jobKey, "of type", job.Type)
    log.Println("Processing order:", variables["orderId"])
    log.Println("Collect money using payment method:", headers["method"])
    
    ctx := context.Background()
    _, err = request.Send(ctx)
    if err != nil {
        panic(err)
    }
    
    log.Println("Successfully completed job")
    close(readyClose)
}

func failJob(client worker.JobClient, job entities.Job) {
    log.Println("Failed to complete job", job.GetKey())
    ctx := context.Background()
    _, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
    if err != nil {
        panic(err)
    }
}
```

* Update the `main()` function:
```go
var readyClose = make(chan struct{})

func main() {
    //...
    
    jobWorker := client.NewJobWorker().JobType("payment-service").Handler(handleJobPaymentService).Open()

    <-readyClose
    jobWorker.Close()
    jobWorker.AwaitClose()
}
```
Run the worker program with the command: `go run main.go`.

You will see output similar to:
```
2022/04/06 16:20:59 Complete job 2251799813685264 of type payment-service
2022/04/06 16:20:59 Processing order: 31243
2022/04/06 16:20:59 Collect money using payment method: VISA
2022/04/06 16:20:59 Successfully completed job
```