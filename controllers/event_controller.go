package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"github.com/zhaojunlucky/mkdocs-cms/services"
)

var eventService = services.NewEventService()

// GetEvents returns all events with pagination and filtering
func GetEvents(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	// Parse filter parameters
	level := models.EventLevel(c.Query("level"))
	source := models.EventSource(c.Query("source"))
	
	var resolved *bool
	if resolvedStr := c.Query("resolved"); resolvedStr != "" {
		resolvedBool, err := strconv.ParseBool(resolvedStr)
		if err == nil {
			resolved = &resolvedBool
		}
	}
	
	events, total, err := eventService.GetAllEvents(page, pageSize, level, source, resolved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	var response []models.EventResponse
	for _, event := range events {
		response = append(response, event.ToResponse(false, false))
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":       response,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetEvent returns a specific event by ID
func GetEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	
	event, err := eventService.GetEventByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	
	c.JSON(http.StatusOK, event.ToResponse(true, true))
}

// GetEventsByResource returns events for a specific resource
func GetEventsByResource(c *gin.Context) {
	resourceType := c.Param("resource_type")
	
	var resourceID *uint
	if resourceIDStr := c.Query("resource_id"); resourceIDStr != "" {
		id, err := strconv.ParseUint(resourceIDStr, 10, 32)
		if err == nil {
			uintID := uint(id)
			resourceID = &uintID
		}
	}
	
	events, err := eventService.GetEventsByResourceType(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	var response []models.EventResponse
	for _, event := range events {
		response = append(response, event.ToResponse(false, false))
	}
	
	c.JSON(http.StatusOK, response)
}

// CreateEvent creates a new event
func CreateEvent(c *gin.Context) {
	var request models.CreateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Capture IP address and user agent if not provided
	if request.IPAddress == "" {
		request.IPAddress = c.ClientIP()
	}
	
	if request.UserAgent == "" {
		request.UserAgent = c.GetHeader("User-Agent")
	}
	
	event, err := eventService.CreateEvent(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, event.ToResponse(true, false))
}

// UpdateEvent updates an existing event
func UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	
	var request models.UpdateEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	event, err := eventService.UpdateEvent(uint(id), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, event.ToResponse(true, false))
}

// DeleteEvent deletes an event
func DeleteEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	
	if err := eventService.DeleteEvent(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
