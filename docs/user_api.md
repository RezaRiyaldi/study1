# API Documentation

## Endpoints
### GET /api/v1/user

**Summary:** RegisterRoutes

**Description:** RegisterRoutes

**Tags:** user 

**Handler:** RegisterRoutes

---
### get /api/v1/users

**Summary:** Get all users

**Description:** Get list of users with pagination and filtering

**Tags:** users 

**Handler:** GetAllUsers

---
### get /api/v1/users/{id}

**Summary:** Get user by ID

**Description:** Get user details by ID

**Tags:** users 

**Handler:** GetUserByID

---
### post /api/v1/users

**Summary:** Create a new user

**Description:** Create a new user with the provided data

**Tags:** users 

**Handler:** CreateUser

---
### put /api/v1/users/{id}

**Summary:** Update user

**Description:** Update user details

**Tags:** users 

**Handler:** UpdateUser

---
### delete /api/v1/users/{id}

**Summary:** Delete user

**Description:** Delete user by ID

**Tags:** users 

**Handler:** DeleteUser

---

## Data Models
### UserHandler

| Field | Type | JSON Tag | Required | Example |
|-------|------|----------|----------|---------|
| service | UserService |  | false |  |
### User

| Field | Type | JSON Tag | Required | Example |
|-------|------|----------|----------|---------|
| ID | uint | id | true | 1 |
| Name | string | name | true | John Doe |
| Email | string | email | true | user@example.com |
| Age | int | age | true | 123 |
| CreatedAt | time.Time | created_at | true | 2023-10-20T10:00:00Z |
| UpdatedAt | time.Time | updated_at | true | 2023-10-20T10:00:00Z |
| DeletedAt | gorm.DeletedAt | - | true | example |
