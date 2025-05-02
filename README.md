# 🔐 Golang JWT Auth API

A RESTful authentication API built using **Golang**, **Gin**, and **MongoDB**. This project demonstrates token-based authentication with **JWT**, secure password handling with **bcrypt**, and cleanly structured Go code for scalable API development.

---

## 🧰 Tech Stack

- **Language:** Golang (Go)
- **Framework:** Gin Gonic
- **Database:** MongoDB (Go Driver)
- **Auth:** JWT, Bcrypt
- **Structure:** MVC with modular folders
- **Tools:** Postman, Docker (optional)

---


---

## 📌 Features

- ✅ JWT-based **user authentication**
- ✅ **Bcrypt** password hashing
- ✅ Middleware to **protect private routes**
- ✅ **MongoDB** integration with native driver
- ✅ Clean code and modular design

---

## 🌐 API Endpoints

| Method | Endpoint             | Description               |
|--------|----------------------|---------------------------|
| POST   | `/users/signup`      | Register a new user       |
| POST   | `/users/login`       | Login and get JWT token   |
| GET    | `/users`             | Get all users (protected) |
| GET    | `/users/:user_id`    | Get user by ID (protected)|

> ⚠️ Protected routes require a JWT token in the `Authorization` header.

---

## 🔑 Environment Variables

Create a `.env` file in the root:

```
PORT=9000
MONGODB_URL=mongodb://127.0.0.1:27017/go-jwt-auth
JWT_SECRET=your_jwt_secret_key
```

---

## 🚀 Getting Started

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Run the App

```bash
go run cmd/main.go
```

---

## 🔐 Using the API

Use **Postman** or any API tool:

### 🔸 Register

- **POST** `/users/signup`
- Body (JSON):

```json
{
  "email": "user@example.com",
  "password": "your_password",
  "first_name": "John",
  "last_name": "Doe"
}
```

### 🔸 Login

- **POST** `/users/login`
- Body (JSON):

```json
{
  "email": "user@example.com",
  "password": "your_password"
}
```

- Response:

```json
{
  "token": "JWT_TOKEN_HERE"
}
```

### 🔸 Authenticated Requests

Include the JWT token in headers:

```
Authorization: Bearer <JWT_TOKEN_HERE>
```

---

## 🧠 Learnings

- Handling authentication securely in Go
- Creating and validating JWTs
- Building middleware for route protection
- Connecting and querying MongoDB from Go
- Designing scalable Go backend APIs

---
