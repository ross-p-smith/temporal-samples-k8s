package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"net/http"
	"os"
	"crypto/tls"
	"crypto/x509"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	// Added as we are not using samples
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/activity"

	//"github.com/temporalio/samples-go/timer"
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

	log.Println("About to execute workflow in background thread.")

	if c == nil {
		log.Fatalln("client is empty")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Unable to execute workflow"))
		return
	}

	we, err := c.ExecuteWorkflow(context.Background(),
		workflowOptions,
		Workflow,
		//timer.SampleTimerWorkflow,
		time.Second*3)

	log.Println("Executed SampleTimerWorkflow.")
		
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

	// c, err = client.NewClient(client.Options
	// 	{
	// 		HostPort: os.Getenv("TEMPORAL_GRPC_ENDPOINT")
	// 	})
	// The client and worker are heavyweight objects that should be created once per process.
	clientOptions, err := ParseClientOptions()
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	if c == nil {
		log.Fatalln("Dial call falled. c is empty")
		return
	}

	w := worker.New(c, "timer", worker.Options{MaxConcurrentActivityExecutionSize: 3})

	// w.RegisterWorkflow(timer.SampleTimerWorkflow)
	// w.RegisterActivity(timer.OrderProcessingActivity)
	// w.RegisterActivity(timer.SendEmailActivity)
	w.RegisterWorkflow(Workflow)
	w.RegisterActivity(Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}

	log.Println("Workflows and Activities are registered.")
}

// ParseClientOptionFlags parses the given arguments into client options. In
// some cases a failure will be returned as an error, in others the process may
// exit with help info.
func ParseClientOptions() (client.Options, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_TLS_CERT"), os.Getenv("TEMPORAL_TLS_KEY"))
	if err != nil {
		return client.Options{}, fmt.Errorf("failed loading client cert and key: %w", err)
	}

	// Load server CA if given
	var serverCAPool *x509.CertPool
	serverCAPool = x509.NewCertPool()
	b, err := os.ReadFile(os.Getenv("TEMPORAL_TLS_CA"))
	if err != nil {
		return client.Options{}, fmt.Errorf("failed reading server CA: %w", err)
	} else if !serverCAPool.AppendCertsFromPEM(b) {
		return client.Options{}, fmt.Errorf("server CA PEM file invalid")
	}

	log.Println("Parsing Client Options.")

	return client.Options{
		HostPort:  os.Getenv("TEMPORAL_ADDRESS"),
		Namespace: "default",
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            serverCAPool,
				ServerName:         os.Getenv("TEMPORAL_TLS_SERVER_NAME"),
				InsecureSkipVerify: true,
			},
		},
	}, nil
}

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {

	log.Println("Entered actual Workflow code.")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
