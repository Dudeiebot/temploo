
# 🍹 temploo

## What's this all about?

Hey there! Welcome to **temploo** - my templates for quickly spinning up new Go projects. This is basically my go-to toolkit that helps me get projects off the ground without reinventing the wheel every time.

## What I'm working with

I've put together a solid stack of tools that play nice together and handle most of what I need in my Go projects:

### 🐘 [PostgreSQL](https://www.postgresql.org)
The reliable database that's powerful enough for serious projects but not overly complex. Perfect for structured data that needs to be queried in flexible ways.

### 🔀 [Chi Router](https://go-chi.io)
A lightweight but feature-rich HTTP router for Go. It's super flexible and keeps my routing clean and organized without unnecessary bloat.

### ⚡ [Redis](https://redis.io)
The lightning-fast in-memory data store that I use for caching, session management, and anywhere I need blazing speed for simple data structures.

### 🔄 [GORM](https://gorm.io)
A fantastic ORM library that makes database operations in Go so much smoother. Helps me work with my Postgres data using Go structs instead of raw SQL.

### ✅ [Go Validator](https://github.com/thedevsaddam/govalidator)
Keeps my data clean and error-free by validating input before it causes problems deeper in the application.

### 📊 [Asynq](https://github.com/hibiken/asynq)
A distributed task queue system for Go that helps me handle background processing and scheduled tasks without headaches.

### 📺 [Asynqmon](https://github.com/hibiken/asynqmon)
A web UI for monitoring and managing those Asynq tasks - super helpful for seeing what's happening with background processes.

## Usage
Add appropriate values to `.env` and run the server

```bash
make dev
```

---

Feel free to clone this repo whenever you need a quick start for a new Go project with this stack!
