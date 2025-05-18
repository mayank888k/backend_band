# Modern Band Booking API

A Go backend server for the Modern Band wedding booking system, using SQLite as the database.

## Features

- OTP generation and verification
- Booking creation and retrieval
- Simulated SMS notifications

## API Endpoints

1. `POST /api/send-otp`
   - Input: `{ "contact_number": "string" }`
   - Generates and "sends" a 6-digit OTP

2. `POST /api/verify-otp`
   - Input: `{ "contact_number": "string", "otp": "string" }`
   - Verifies the OTP

3. `POST /api/book`
   - Input: All booking fields
   - Creates a booking (requires phone verification)

4. `GET /api/booking?booking_id=X` or `GET /api/booking?contact_number=Y`
   - Retrieves booking(s) by ID or phone number

## Setup Instructions

### Prerequisites

- Go 1.18 or higher
- SQLite3

### Installation

1. Clone the repository:
```bash
git clone https://github.com/your-username/modernband.git
cd modernband/backend
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o booking-api ./cmd/api
```

4. Run the server:
```bash
./booking-api
```

The API will be available at `http://localhost:8080`.

## Environment Variables

Create a `.env` file with the following variables (optional):

```
PORT=8080
```

## Integration with Frontend

The API is designed to integrate with the React frontend. The frontend makes API calls to:

1. Send OTP when the "Book Now" button is clicked
2. Verify OTP with the entered code
3. Submit the booking form data
4. Retrieve booking details for reference

## Database

The application uses SQLite, which creates a database file at `./data/bookings.db`. The database is automatically initialized with the required tables on first run. 