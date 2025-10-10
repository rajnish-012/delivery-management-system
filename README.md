# 🚚 Delivery Management System

A backend service built using **Go (Golang)** for managing delivery operations — including users, orders, and real-time tracking.  
The project follows a **clean architecture** with modular components and supports containerized deployment using **Docker**.

---

## 🧩 Features

- 🔐 User authentication with JWT  
- 📦 Order creation, tracking, and management  
- 🗄️ PostgreSQL integration for data storage  
- ⚡ Redis caching for performance boost  
- 🧱 Modular and maintainable Go code structure  
- 🐳 Ready-to-run Docker setup  

---

## 📁 Project Structure

delivery-management-system/
├── cmd/
│ └── server/
│ └── main.go # Application entry point
├── internal/
│ ├── api/ # HTTP request handlers
│ ├── auth/ # JWT authentication logic
│ ├── database/ # PostgreSQL & Redis connections
│ ├── models/ # Data models for users and orders
│ ├── orders/ # Order management logic
│ └── tests/ # Unit tests
├── migrations/ # Database schema setup
├── docker-compose.yml # Docker configuration
├── go.mod # Go module dependencies
├── go.sum
└── api_test.http # Example API testing file


---

## ⚙️ Installation and Setup

### 🧱 Prerequisites
Make sure you have the following installed:
- [Go](https://go.dev/dl/) 1.21+
- [Docker](https://www.docker.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)

---

### 📥 Clone the Repository

- git clone https://github.com/rajnish-012/delivery-management-system.git
- cd delivery-management-system

## ⚙️Setup Environment Variables

Create a .env file in the root directory and configure it as below:

POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
POSTGRES_DB=delivery_db
POSTGRES_HOST=localhost
POSTGRES_PORT=5432

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your_secret_key


## 🐳 Run with Docker

Use Docker Compose to build and start all services:

docker-compose up --build

## Running Tests

To run all test cases:
go test ./internal/tests/...


## 🗃️ Database Migration

Initialize the database schema:

psql -U <user> -d delivery_db -f migrations/0001_init.sql

## 🧰 Tech Stack
Component	Technology
Language	Go (Golang)
Database	PostgreSQL
Cache	Redis
Authentication	JWT
Containerization	Docker & Docker Compose

## 🧑‍💻 Author

Rajnish Kumar

## 📄 License

This project is licensed under the MIT License.
You are free to use, modify, and distribute this software as long as proper credit is given.
