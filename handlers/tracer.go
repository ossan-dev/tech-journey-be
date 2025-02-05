package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/trace"
)

type FlightRecorderTracer struct {
	FlightRecorderTracer *trace.FlightRecorder
}

func NewFlightRecorder() *FlightRecorderTracer {
	return &FlightRecorderTracer{
		FlightRecorderTracer: trace.NewFlightRecorder(),
	}
}

func (f *FlightRecorderTracer) Trace(c *gin.Context) {
	file, err := os.Create("cmd/flight_recorder.out")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()
	if _, err := f.FlightRecorderTracer.WriteTo(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
