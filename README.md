# bookapi

A RESTful API for managing books built with Go and PostgreSQL,
seeded with real data from the Open Library API.

## Main Idea

Understanding how APIs expose data and how servers and clients
interact with each other through APIs.

## Stack

- Go + Gin
- PostgreSQL
- Python (scraper)

## What I Learned

**New concepts**

- Authentication and Authorization: `bcrypt` for password hashing,
  `rand` for 32-byte hex session tokens, bearer token auth
- Middleware: conditions that allow or block handler functions,
  used for auth checks, role checks and rate limits
- PostgreSQL: different syntax from SQLite. `RETURNING id` instead
  of `LastInsertId()`, `$1 $2 $3` placeholders instead of `?`
- Dynamic SQL: build queries conditionally using `fmt.Sprintf`,
  `[]any` args slice and `WHERE 1=1`
- Concurrency: goroutines run requests concurrently. race
  conditions on shared data require `sync.Mutex` to lock and
  unlock access
- Rate limiting: token bucket per IP address per route group.
  each IP gets its own `rate.Limiter` stored in `map[string]*rate.Limiter`
- Database indexing: used `EXPLAIN ANALYZE` to verify index usage.
  execution time dropped from 0.968ms to 0.360ms on a 2000 row table

**Reinforced concepts**

- CRUD: routes, error handling, request parsing in Go/Gin
- Raw SQL over ORM: to understand what happens under the hood
  before reaching for abstractions

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL

### Setup

1. Clone the repo `git clone https://github.com/WaiYanKyaw95/books-api.git`

3. Create a PostgreSQL database `createdb books`

4. Run the server, tables are created automatically `go run .`

5. Seed the database
   `go run seed/main.go`

Server runs on `http://localhost:8080`

## API Endpoints

### Public
- `GET /books`       list books with pagination, filtering and sorting
- `GET /books/:id`   single book by id

### Auth
- `POST /register`    create account
- `POST /login`       returns bearer token
- `POST /logout`      invalidate token, requires auth

### Admin only
- `POST /books`       create book
- `PUT /books/:id`   update book
- `DELETE /books/:id`   delete book

## Query Parameters

`GET /books` supports:

  - page      page number, default 1
  - limit     results per page, default 10
  - author    filter by author name
  - year      filter by publication year
  - subject   filter by subject, case insensitive
  - sort      sort by id, title or year, default id
  - order     asc or desc, default asc

## Authentication

- Register and login to get a bearer token.
- Include it in every protected request: `Authorization: Bearer <token>`

## Rate Limits

  - Public routes    100 requests per minute
  - Auth routes      10 requests per minute
  - Protected        10 requests per minute
  - Admin routes     30 requests per minute

## Final Note

This project helped me work through backend fundamentals
alongside concepts I had never touched before. Testing is not
included here. The plan is to rebuild this in Python with
FastAPI and approach it with TDD from the start.
