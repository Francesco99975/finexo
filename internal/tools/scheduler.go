package tools

import (
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/robfig/cron/v3"
)

type Job struct {
	ID      string
	EntryID cron.EntryID
}

var jobMutex sync.Mutex
var cronScheduler *cron.Cron
var jobs = make(map[string]Job)

func init() {
	// This function will be called automatically before main() starts
	log.Info("Initializing cron scheduler...")
	cronScheduler = cron.New()
	cronScheduler.Start() // Start the cron scheduler
}

// AddJob schedules a new job and stores it in the jobs map
func AddJob(id string, schedule string, task func()) error {
	jobMutex.Lock()
	defer jobMutex.Unlock()

	// Schedule the task
	entryID, err := cronScheduler.AddFunc(schedule, task)
	if err != nil {
		return err
	}

	//Check if the job with the same ID already exists
	if _, ok := jobs[id]; ok {
		log.Warnf("Job with ID: %s already exists\n", id)
		return nil
	}

	// Store the job with its entry ID
	jobs[id] = Job{
		ID:      id,
		EntryID: entryID,
	}

	log.Infof("Scheduled job with ID: %s\n", id)
	return nil
}

func UpdateJob(id string, schedule string, task func()) error {
	jobMutex.Lock()
	defer jobMutex.Unlock()

	if job, ok := jobs[id]; ok {
		cronScheduler.Remove(job.EntryID)
		entryID, err := cronScheduler.AddFunc(schedule, task)
		if err != nil {
			return err
		}
		jobs[id] = Job{
			ID:      id,
			EntryID: entryID,
		}
		log.Infof("Updated job with ID: %d\n", id)
	} else {
		log.Errorf("Job with ID: %d not found\n", id)
	}

	return nil
}

// RemoveJob deletes a job by its ID
func RemoveJob(id string) {
	jobMutex.Lock()
	defer jobMutex.Unlock()

	if job, ok := jobs[id]; ok {
		cronScheduler.Remove(job.EntryID)
		delete(jobs, id)
		log.Infof("Removed job with ID: %d\n", id)
	} else {
		log.Errorf("Job with ID: %d not found\n", id)
	}
}
