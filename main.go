package main

import (
	"context"
	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"
	"github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	"github.com/camunda/zeebe/clients/go/v8/pkg/worker"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var readyClose = make(chan struct{})

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

	jobWorker := client.NewJobWorker().JobType("get-time").Handler(HandleJob).Open()

	<-readyClose
	jobWorker.Close()
	jobWorker.AwaitClose()

}

func DeployResource(ctx context.Context, client zbc.Client) error {
	response, err := client.NewDeployResourceCommand().AddResourceFile("resources/order-process.bpmn").Send(ctx)
	if err != nil {
		return err
	}
	log.Println(response.String())
	return nil
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

func HandleJob(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	headers, err := job.GetCustomHeadersAsMap()
	if err != nil {
		// failed to handle job as we require the custom job headers
		if err := failJob(client, job); err != nil {
			log.Println("Error when fail job: " + err.Error())
		}
		return
	}

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		if err := failJob(client, job); err != nil {
			log.Println("Error when fail job: " + err.Error())
		}
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)
	log.Println("Processing order:", variables["orderId"])
	log.Println("Collect money using payment method:", headers["method"])

	ctx := context.Background()
	_, err = client.NewCompleteJobCommand().JobKey(job.Key).Send(ctx)
	if err != nil {
		return
	}

	log.Println("Successfully completed job")
	close(readyClose)
}

func failJob(client worker.JobClient, job entities.Job) error {
	log.Println("Failed to complete job", job.GetKey())
	ctx := context.Background()
	_, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
	if err != nil {
		return err
	}

	return nil
}
