package worker

import (
	"time"

	"github.com/calvinmclean/automated-garden/garden-app/pkg/influxdb"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/mqtt"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/storage"
	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	scheduleJobsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "garden_app",
		Name:      "scheduled_jobs",
		Help:      "gauge of the currently-scheduled jobs",
	}, []string{"type", "id"})
	schedulerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "garden_app",
		Name:      "scheduler_errors",
		Help:      "count of errors that occur in the background and do not have any visibility except logs",
	}, []string{"type", "id"})
)

// Worker contains the necessary clients to schedule and execute actions
type Worker struct {
	storageClient  *storage.Client
	influxdbClient influxdb.Client
	mqttClient     mqtt.Client
	scheduler      *gocron.Scheduler
	logger         *logrus.Entry
}

// NewWorker creates a Worker with specified clients
func NewWorker(
	storageClient *storage.Client,
	influxdbClient influxdb.Client,
	mqttClient mqtt.Client,
	logger *logrus.Logger,
) *Worker {
	return &Worker{
		storageClient:  storageClient,
		influxdbClient: influxdbClient,
		mqttClient:     mqttClient,
		scheduler:      gocron.NewScheduler(time.Local),
		logger:         logger.WithField("source", "worker"),
	}
}

// StartAsync starts the Worker's background jobs
func (w *Worker) StartAsync() {
	w.scheduler.StartAsync()
	prometheus.MustRegister(
		scheduleJobsGauge,
		schedulerErrors,
	)
}

// Stop stops the Worker's background jobs
func (w *Worker) Stop() {
	w.scheduler.Stop()
	if w.mqttClient != nil {
		w.mqttClient.Disconnect(100)
	}
	if w.influxdbClient != nil {
		w.influxdbClient.Close()
	}

	prometheus.Unregister(scheduleJobsGauge)
	prometheus.Unregister(schedulerErrors)
}
