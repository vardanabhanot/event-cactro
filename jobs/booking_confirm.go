package jobs

import (
	"log"
	"time"
)

// BookingConfirmJob simulates sending a booking confirmation email to a customer.
type BookingConfirmJob struct {
	CustomerName  string
	CustomerEmail string
	EventTitle    string
	EventDate     time.Time
	Tickets       int
}

func (j BookingConfirmJob) Process() {
	log.Printf("[JOB] BookingConfirm → Sending confirmation email to %s <%s>", j.CustomerName, j.CustomerEmail)
	log.Printf("[JOB] BookingConfirm → Subject: Your booking for \"%s\" on %s is confirmed (%d ticket(s))",
		j.EventTitle,
		j.EventDate.Format("Jan 02, 2006 15:04"),
		j.Tickets,
	)
}
