package jobs

import "log"

// Recipient holds the minimal info needed to notify a customer.
type Recipient struct {
	Name  string
	Email string
}

// EventNotifyJob simulates notifying all customers booked for an updated event.
type EventNotifyJob struct {
	EventTitle string
	Recipients []Recipient
}

func (j EventNotifyJob) Process() {
	log.Printf("[JOB] EventNotify → Event \"%s\" was updated. Notifying %d customer(s)...", j.EventTitle, len(j.Recipients))
	for _, r := range j.Recipients {
		log.Printf("[JOB] EventNotify → Notifying %s <%s>: The event \"%s\" you booked has been updated.",
			r.Name, r.Email, j.EventTitle,
		)
	}
}
