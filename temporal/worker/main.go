package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"net/http"
	"os"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/timer"
)

var c client.Client

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong")
}

func start(w http.ResponseWriter, r *http.Request) {

	workflowOptions := client.StartWorkflowOptions{
		ID:        "timer_" + uuid.New(),
		TaskQueue: "timer",
	}

	we, err := c.ExecuteWorkflow(context.Background(),
		workflowOptions,
		timer.SampleTimerWorkflow,
		time.Second*3)
		
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Unable to execute workflow"))
		return
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	w.WriteHeader(http.StatusAccepted)
	result := fmt.Sprintf("Started workflow ID=%s, RunID=%s", we.GetID(), we.GetRunID())
	w.Write([]byte(result))
}

func main() {
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/delay", start)

	// Start Server
	go func() {
		log.Println("Starting Web Server")
		http.ListenAndServe(":8080", nil)
	}()

	// The client and worker are heavyweight objects that should be created once per process.
	var err error
	c, err = client.NewClient(client.Options{HostPort: os.Getenv("TEMPORAL_GRPC_ENDPOINT")})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "timer", worker.Options{MaxConcurrentActivityExecutionSize: 3})

	w.RegisterWorkflow(timer.SampleTimerWorkflow)
	w.RegisterActivity(timer.OrderProcessingActivity)
	w.RegisterActivity(timer.SendEmailActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
