# Booking Management System for Cactro
This is a event booking APP API build with Go, it's endpoints can be accessed at /help

# Language Choice
I had 2 options on which I could have made this, one being PHP and second Go.
I did not go with PHP becuase PHP does not have Async, there are ways to make it works, but it could be messy, where as Go has concurrency build in.
Other point is as PHP is synchronous, and to check if there is a job in the queue would have to run a cron job every minute or so, which is not the case in Go, as Go is a long running process and we can use goroutines to handle the jobs.

NOTE: Laravel has queue system but Go implementation is lighter.

# Database Choice
For the Job queue I thought of 2 options one being Redis and Second being SQLite, I choose SQLite because it was easy to setup on any environment. Although it has File Write Locks and can be a bottleneck if concurrency increases but in this case it is save to use it.

The current structure is MVC like.

## API Endpoints

- **GET /help**: Get all the API documentation.
- **POST /api/auth/register**: Register a new user.
- **POST /api/auth/login**: Login with email and password.
- **GET /api/events**: Get all events.
- **GET /api/events/{id}**: Get a single event by ID.
- **POST /api/events**: Create a new event.
- **PUT /api/events/{id}**: Update an event.
- **DELETE /api/events/{id}**: Delete an event.
- **POST /api/bookings**: Create a new booking.
- **GET /api/bookings**: Get all bookings.
- **GET /api/bookings/{id}**: Get a single booking by ID.
- **DELETE /api/bookings/{id}**: Delete a booking.

