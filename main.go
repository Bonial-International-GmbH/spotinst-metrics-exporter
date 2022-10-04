package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bonial-International-GmbH/spotinst-metrics-exporter/pkg/collectors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spotinst/spotinst-sdk-go/service/mcs"
	"github.com/spotinst/spotinst-sdk-go/service/ocean"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	sess := session.New()
	mcsClient := mcs.New(sess)

	oceanClient := ocean.New(sess)

	clusters, err := getOceanAWSClusters(ctx, oceanClient)
	if err != nil {
		log.Fatal(err)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewOceanAWSClusterCostsCollector(ctx, mcsClient, clusters))
	registry.MustRegister(collectors.NewOceanAWSRightSizingCollector(ctx, oceanClient.CloudProviderAWS(), clusters))

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}))

	if err := listenAndServe(ctx, handler, *addr); err != nil {
		log.Fatal(err)
	}
}

func handleSignals(cancelFunc func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	<-signals
	log.Println("received signal, terminating...")
	cancelFunc()
}

func listenAndServe(ctx context.Context, handler http.Handler, addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("listening on %s", addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	<-ctx.Done()

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func getOceanAWSClusters(ctx context.Context, client ocean.Service) ([]*aws.Cluster, error) {
	output, err := client.CloudProviderAWS().ListClusters(ctx, &aws.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	return output.Clusters, nil
}
