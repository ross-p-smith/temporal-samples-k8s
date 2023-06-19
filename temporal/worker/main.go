package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"net/http"
	"os"

	// Added for TLS
	"crypto/tls"
	"crypto/x509"

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

	clientOptions, err := ParseClientOptions()
	if err != nil {
		log.Fatalf("Unable to create client connection configuration: %v", err)
	}

	c, err = client.NewClient(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client.", err)
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

func ParseClientOptions() (client.Options, error) {
	cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_TLS_CERT"), os.Getenv("TEMPORAL_TLS_KEY"))
	if err != nil {
		return client.Options{}, fmt.Errorf("failed loading client cert and key: %w", err)
	}

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
		HostPort:  os.Getenv("TEMPORAL_GRPC_ENDPOINT"),
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
