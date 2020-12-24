package services

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/wufe/monito/worker/models"
)

type WorkingQueue struct {
	*models.Queue
	db           *gorm.DB
	orchestrator *QueueOrchestrator
}

func NewWorkingQueue(orchestrator *QueueOrchestrator, queue *models.Queue, db *gorm.DB) *WorkingQueue {
	return &WorkingQueue{
		Queue:        queue,
		db:           db,
		orchestrator: orchestrator,
	}
}

// func (workingQueue *WorkingQueue) Start() {
func (workingQueue *WorkingQueue) Work(request *models.Request) {
	// requestType := workingQueue.getRequestTypeByQueue()

	// request := <-workingQueue.orchestrator.GetRequest(requestType)
	fmt.Println(fmt.Sprintf("Queue %s is processing the request %s.", workingQueue.UUID, request.UUID))
	workingQueue.updateRequest(request)
	workingQueue.startProcessingRequest(request)
	fmt.Println(fmt.Sprintf("Queue %s finished processing the request %s.", workingQueue.UUID, request.UUID))
	fmt.Println("############################################################")

	// utils.SetTimeout(workingQueue.Start, 1000)

}

func (workingQueue *WorkingQueue) updateRequest(request *models.Request) {
	request.Status = models.RequestStatusInProgress
	request.UpdatedAt = time.Now()
	workingQueue.db.Model(&models.Request{}).
		Update(request)
}

func (workingQueue *WorkingQueue) startProcessingRequest(request *models.Request) {
	NewJob(request, workingQueue.db).Start()
}

func (workingQueue *WorkingQueue) GetRequestTypeByQueue() models.RequestType {
	// func (workingQueue *WorkingQueue) getRequestTypeByQueue() models.RequestType {
	var requestType models.RequestType

	switch workingQueue.Queue.Type {
	case models.QueueTypeBasic:
		requestType = models.RequestTypeBatch
	case models.QueueTypePriority:
		requestType = models.RequestTypeBatch
	case models.QueueTypeSimple:
		requestType = models.RequestTypeSimple
	}

	return requestType
}

func (workingQueue *WorkingQueue) queueTypeLabel() string {
	var queueTypeLabel string
	switch workingQueue.Queue.Type {
	case models.QueueTypeSimple:
		queueTypeLabel = "simple"
	case models.QueueTypeBasic:
		queueTypeLabel = "basic"
	case models.QueueTypePriority:
		queueTypeLabel = "priority"
	}
	return queueTypeLabel
}

func (workingQueue *WorkingQueue) log(text string) {
	fmt.Println(fmt.Sprintf("[Queue #%d:%s] %s", workingQueue.Queue.ID, workingQueue.queueTypeLabel(), text))
}
