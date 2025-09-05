package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

// TeslaMateAPICarsStatusV1 provides car status directly from PostgreSQL (clean version)
func TeslaMateAPICarsStatusV1(c *gin.Context) {
	// Parse car ID from URL
	carID := convertStringToInteger(c.Param("CarID"))

	// Initialize services
	statusService := NewCarStatusService(db)
	mapper := NewCarStatusMapper()

	// Get car status from database
	statusData, err := statusService.GetCarStatus(carID)
	if err != nil {
		TeslaMateAPIHandleErrorResponse(c, "TeslaMateAPICarsStatusV1", "Failed to retrieve car status", err.Error())
		return
	}

	// Determine vehicle state
	vehicleState := statusService.DetermineVehicleState(statusData)

	// Map database results to API response format
	response := mapper.MapToResponse(statusData, vehicleState)

	// Apply unit conversions based on user preferences
	mapper.ApplyUnitConversions(response)

	// Log charging status for debugging
	if gin.IsDebugging() {
		log.Printf("[debug] TeslaMateAPICarsStatusV1 - Car %d: plugged_in=%t, is_charging_from_db=%t", 
			carID, 
			response.Status.ChargingDetails.PluggedIn,
			statusData.IsCharging.Valid && statusData.IsCharging.Bool,
		)
	}

	TeslaMateAPIHandleSuccessResponse(c, "TeslaMateAPICarsStatusV1", response)
}