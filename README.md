# Go Gin API Project

This is a sample API project built with Go and the Gin framework. It includes user authentication and authorization, with options for password-based, SMS-based, and email-based login. The project is containerized with Docker and includes Swagger documentation.

## Features

- **Go + Gin**: A fast and lightweight framework for building APIs.
- **PostgreSQL**: A powerful open-source relational database.
- **GORM**: A developer-friendly ORM for Go.
- **JWT Authentication**: Secure your API with JSON Web Tokens.
- **Role-Based Authorization**: Control access to endpoints based on user roles (admin, user).
- **Multiple Login Methods**:
  - Phone number and password
  - Phone number and SMS code
  - Email and verification code
- **Persian Phone Number Validation**: Ensures that phone numbers are in the correct format.
- **Dockerized**: Run the entire application and database with a single command.
- **Swagger Documentation**: Interactive API documentation.

## Prerequisites

- Go (1.21 or later)
- Docker
- Docker Compose

## Getting Started

### Running with Docker (Recommended)

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd <repository-directory>
    ```

2.  **Create a `.env` file:**
    Copy the example `.env.example` to `.env` and fill in the required environment variables.
    ```sh
    cp .env.example .env
    ```

3.  **Build and run the application:**
    ```sh
    docker-compose up --build
    ```

The application will be available at `http://localhost:8080`.

### Running Locally

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd <repository-directory>
    ```

2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Set up the database:**
    Make sure you have a PostgreSQL instance running. You can use Docker to start one:
    ```sh
    docker run --name my-postgres -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=mydatabase -p 5432:5432 -d postgres
    ```

4.  **Set environment variables:**
    Create a `.env` file or export the following environment variables:
    ```sh
    export DB_HOST=localhost
    export DB_USER=user
    export DB_PASSWORD=password
    export DB_NAME=mydatabase
    export DB_PORT=5432
    export JWT_SECRET=a-very-secret-key
    ```

5.  **Run the application:**
    ```sh
    go run cmd/server/main.go
    ```

The application will be available at `http://localhost:8080`.

## API Documentation

Once the application is running, you can access the Swagger documentation at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## API Endpoints

### Authentication

- `POST /signup`: Create a new user.
- `POST /login`: Log in with phone number and password.
- `POST /login/sms/request`: Request an SMS verification code.
- `POST /login/sms/verify`: Verify the SMS code and get a JWT.
- `POST /login/email/request`: Request an email verification code.
- `POST /login/email/verify`: Verify the email code and get a JWT.

### User Management (Admin only)

- `GET /api/v1/users`: Get a list of all users.
- `GET /api/v1/users/{id}`: Get a single user by ID.
- `PUT /api/v1/users/{id}`: Update a user's information.
- `DELETE /api/v1/users/{id}`: Delete a user.
- `PUT /api/v1/users/{id}/role`: Assign a role to a user.

### Restaurant Management

- `POST /api/v1/restaurants`: Create a new restaurant (requires authentication).
- `GET /api/v1/restaurants/{id}`: Get a restaurant's details.
- `PUT /api/v1/restaurants/{id}`: Update a restaurant (owner or admin only).
- `DELETE /api/v1/restaurants/{id}`: Delete a restaurant (owner or admin only).
- `GET /api/v1/restaurants/{id}/qrcode`: Get a QR code for the restaurant's menu.
- `PUT /api/v1/restaurants/{id}/daily-menu`: Set the daily menu for a restaurant (owner or admin only).

### Menu Management

- `POST /api/v1/restaurants/{id}/menus`: Create a new menu for a restaurant (owner or admin only).
- `GET /api/v1/menus/{menu_id}`: Get a menu's details.
- `PUT /api/v1/menus/{menu_id}`: Update a menu (owner or admin only).
- `DELETE /api/v1/menus/{menu_id}`: Delete a menu (owner or admin only).
- `POST /api/v1/menus/{menu_id}/items`: Add a new item to a menu (owner or admin only).

### Menu Item Management

- `PUT /api/v1/menu-items/{item_id}`: Update a menu item (owner or admin only).
- `DELETE /api/v1/menu-items/{item_id}`: Delete a menu item (owner or admin only).
