package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/wufe/monito/worker/models"
	"github.com/wufe/monito/worker/utils"
)

type JobProcess struct {
	request           *models.Request
	options           *models.RequestOptions
	db                *gorm.DB
	retrievalDoneChan *chan struct{}
	fetchDoneChan     *chan struct{}
	stopChan          *chan struct{}
	linksChan         chan *models.Link
}

func NewJob(request *models.Request, db *gorm.DB) *JobProcess {
	return &JobProcess{
		request: request,
		db:      db,
	}
}

func (job *JobProcess) Start() {
	job.resetLinks()
	job.loadOptions()

	stopChan := make(chan struct{}, 1)
	job.stopChan = &stopChan
	retrievalDoneChan := make(chan struct{}, 1)
	job.retrievalDoneChan = &retrievalDoneChan
	fetchDoneChan := make(chan struct{}, 1)
	job.fetchDoneChan = &fetchDoneChan

	job.startStatusCheckProcess()
	job.startLinkRetrievalProcess()
	job.startFetch()
}

// Recover from inconsistent state
func (job *JobProcess) resetLinks() {
	job.db.Model(&models.Link{}).
		Where(&models.Link{RequestID: job.request.ID, Status: models.LinkStatusAcknowledged}).
		Update(map[string]interface{}{"status": models.LinkStatusIdle, "updated_at": time.Now()})
}

func (job *JobProcess) loadOptions() {

	requestOptions := models.RequestOptions{
		Method:    models.RequestOptionsMethodTypeHead,
		Redirects: 5,
		Threads:   2,
		Timeout:   10000,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36",
	}
	err := json.Unmarshal([]byte(job.request.Options), &requestOptions)
	if err != nil {
		fmt.Println("Could not deserialize request options: going with default values.")
	}
	job.options = &requestOptions
}

func (job *JobProcess) startStatusCheckProcess() {
	go func() {
	L:
		for {
			select {
			case <-*job.retrievalDoneChan:
				fmt.Println("#001 Status check process ending because there are no more links to be requested..")
				*job.retrievalDoneChan <- struct{}{}
				fmt.Println("#001 Done.")
				break L
			case <-*job.fetchDoneChan:
				fmt.Println("#002 Status check process ending because the fetch process has done..")
				*job.fetchDoneChan <- struct{}{}
				fmt.Println("#002 Done.")
				break L
			default:
				status := job.getJobStatus()
				if status == models.RequestStatusAbort {
					fmt.Println("#003 Status check process ending because abortion has been requested..")
					*job.stopChan <- struct{}{}
					fmt.Println("#003 Done.")
					break L
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

func (job *JobProcess) getJobStatus() models.RequestStatus {

	foundRequest := models.Request{}

	job.db.Model(&models.Request{}).
		Where(&models.Request{ID: job.request.ID}).
		First(&foundRequest)

	job.request.Status = foundRequest.Status

	return foundRequest.Status
}

func (job *JobProcess) startLinkRetrievalProcess() {
	job.linksChan = make(chan *models.Link, job.options.Threads)
	go func() {
		job.retrievalProcess()
	}()
}

func (job *JobProcess) retrievalProcess() {
	var lastLinkID uint
L:
	for {
		select {
		case <-*job.stopChan:
			*job.stopChan <- struct{}{}
			fmt.Println("Retrieval process aborted.")
			break L
		default:
			link := job.getLink(lastLinkID)
			if link.ID > 0 {
				lastLinkID = link.ID
				job.linksChan <- link
			} else {
				*job.retrievalDoneChan <- struct{}{} // <- 1
				break L
			}
		}
	}

	job.request.Status = models.RequestStatusDone

	job.db.Model(&models.Request{}).
		Update(job.request)

	fmt.Println("Retrieval queue process ended.")
}

func (job *JobProcess) getLink(greaterThanID uint) *models.Link {

	foundLink := models.Link{}

	job.db.Model(&models.Link{}).
		Where("request_id = ? AND status = ? AND id > ?", job.request.ID, models.LinkStatusIdle, greaterThanID).
		First(&foundLink)

	if foundLink.ID > 0 {
		job.db.Model(&foundLink).
			Where(&models.Link{ID: foundLink.ID}).
			Update(map[string]interface{}{"status": models.LinkStatusAcknowledged, "updated_at": time.Now()})
	}

	return &foundLink
}

func (job *JobProcess) startFetch() {

	requestQueueChannel := make(chan struct{}, job.options.Threads)
	for i := 0; i < job.options.Threads; i++ {
		requestQueueChannel <- struct{}{}
	}

	var wg sync.WaitGroup

L:
	for {
		select {
		case <-requestQueueChannel:
			go func() {
				wg.Add(1)
				<-job.performLinkRequest()
				requestQueueChannel <- struct{}{}
				wg.Done()
			}()
		case <-*job.stopChan:
			*job.stopChan <- struct{}{}
			fmt.Println("Fetch process aborted")
			break L
		// Should enter here only if request queue channel is empty and done
		case <-*job.retrievalDoneChan:
			fmt.Println("#004 Fetch process cannot fetch more links..")
			*job.retrievalDoneChan <- struct{}{}
			fmt.Println("#004 Stopping fetch process..")
			*job.fetchDoneChan <- struct{}{}
			fmt.Println("#004 Done")
			break L
		default:
		}
	}

	wg.Wait()
	fmt.Println(fmt.Sprintf("Process ended for request #%d", job.request.ID))
}

func (job *JobProcess) performLinkRequest() <-chan struct{} {
	doneChan := make(chan struct{}, 1)
	go func() {
		select {
		// If there are no links to be retrieved
		case <-*job.retrievalDoneChan:
			// anyway, restore the situation
			*job.retrievalDoneChan <- struct{}{}
			doneChan <- struct{}{}
		case link := <-job.linksChan:
			<-NewHTTPRequestService(link, job).Send()
			utils.SetTimeout(func() {
				doneChan <- struct{}{}
			}, 1000)
		}
	}()
	return doneChan
}
