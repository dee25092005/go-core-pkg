# go-core-pkg

A small collection of shared Go packages I use across my backend projects (built with [Echo](https://echo.labstack.com/)).
Instead of copy-pasting the same error handling, auth, logging, storage, and payment code into every new project, I put it here once and import it everywhere.

```
go get github.com/dee25092005/go-core-pkg
```

----

## What's inside

| Package | What it does |
|---|---|
| [`apperrors`](#apperrors) | Consistent app errors (404, 400, 401, etc.) + validation errors |
| [`auth`](#auth) | Create and check JWT login tokens |
| [`database`](#database) | Connect to Postgres with a ready-to-use connection pool |
| [`middleware`](#middleware) | Echo middlewares: auth check, permission check, error handler, request logger |
| [`phajay`](#phajay) | Client for PhaJay (Lao payment gateway) — QR code payments |
| [`response`](#response) | Simple success response shape |
| [`storage`](#storage) | Upload/delete files on Cloudflare R2 (S3-compatible) |
| [`utils`](#utils) | Small helpers: random token, hash token |

---

## `apperrors`

**What it is for:** A standard way to create errors that always look the same when sent back to the client as JSON, like:

```json
{
  "error": {
    "code": 404,
    "status": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

**How to use it:**

```go
import "github.com/dee25092005/go-core-pkg/apperrors"

// Return a 404 error
return apperrors.NotFound("user not found")

// Return a 400 error
return apperrors.BadRequest("email is required")

// Return a 401 error
return apperrors.Unauthorized("invalid token")

// Return a 403 error
return apperrors.Forbidden("you cannot access this")

// Return a 409 error (already exists)
return apperrors.Conflict("email already used")

// Return a 500 error (wrap a real Go error)
return apperrors.Internal(err)
```

**Field validation errors** (when a form/request has multiple bad fields):

This follows the same idea as [Google's API error design](https://cloud.google.com/apis/design/errors#error_model) — instead of stopping at the *first* bad field, you collect *every* bad field first, then return them all together in one `detail` array. This way the client gets the full list of what's wrong in one request, instead of fixing one field, resubmitting, hitting the next error, fixing that, and so on.

```go
import "github.com/dee25092005/go-core-pkg/apperrors"

func handler(c echo.Context) error {
    validation := apperrors.Validation()

    fileID := c.Param("id")
    if fileID == "" {
        validation.Add(
            "id",              // field: which field is wrong
            "required",        // reason: short machine-readable reason
            "id is required",  // error: human-readable message
        )
    }

    name := c.FormValue("name")
    if name == "" {
        validation.Add("name", "required", "name is required")
    }

    // check as many fields as you want — each bad one gets added

    if err := validation.Error(); err != nil {
        return err
    }

    // all fields are OK, continue with the handler...
    return c.JSON(200, response.SuccessResponse{Message: "ok"})
}
```

`validation.Error()` returns `nil` if nothing was added, so it's always safe to call it once at the end and just `return err` if it's not `nil`.

When this hits `middleware.CustomHTTPErrorHandler`, it becomes one JSON response listing every bad field at once:

```json
{
  "error": {
    "code": 400,
    "status": "INVALID_ARGUMENT",
    "message": "Validation failed",
    "detail": [
      { "field": "id", "reason": "required", "error": "id is required" },
      { "field": "name", "reason": "required", "error": "name is required" }
    ]
  }
}
```

**When to use `apperrors.Validation()` vs. the single-error helpers (`apperrors.BadRequest`, etc.):**

| Situation | Use |
|---|---|
| Checking several fields on one request (a form, a create/update body) | `apperrors.Validation()` — collects every bad field, one response |
| One single problem, nothing to "collect" (e.g. "user not found", "not allowed") | The direct helpers: `apperrors.NotFound(...)`, `apperrors.Forbidden(...)`, `apperrors.Conflict(...)`, etc. |
| An unexpected Go error (DB failed, file system error) | `apperrors.Internal(err)` |

You don't need to build the JSON yourself — just `return` the error from your Echo handler, and the `middleware.CustomHTTPErrorHandler` (see below) turns it into the right JSON response automatically.

---

## `auth`

**What it is for:** Making and checking JWT tokens (used for login sessions).

**How to use it:**

```go
import (
    "time"
    "github.com/dee25092005/go-core-pkg/auth"
)

// 1. When a user logs in, create a token for them
token, err := auth.GenerateJWT(user.ID, user.Tier, "my-secret-key", 24*time.Hour)
if err != nil {
    // handle error
}
// send `token` back to the user

// 2. Later, when a request comes in with that token, check it
claims, err := auth.VerifyJWT(token, "my-secret-key")
if err != nil {
    // token is invalid or expired
}
fmt.Println(claims.UserID, claims.Tier)
```

In most cases you won't call this directly — the `middleware.AuthMiddleware` below does step 2 for you on every request.

**Checking a Google login (OIDC)** — for "Sign in with Google" style login, where the frontend gives you a Google ID token and you need to check it's real:

```go
import "github.com/dee25092005/go-core-pkg/auth"

// 1. Set this up once when your app starts (it caches Google's public keys)
validator, err := auth.NewOIDCValidator("https://www.googleapis.com/oauth2/v3/certs")
if err != nil {
    // handle error
}

// 2. When a request comes in with a Google ID token, check it
claims, err := validator.VerifyGoogleIDToken(idToken, "your-google-client-id.apps.googleusercontent.com")
if err != nil {
    // token is invalid, expired, or not meant for your app
}

fmt.Println(claims.Email, claims.EmailVerified)
```

What it checks for you:
- The token is really signed by Google (using Google's public keys, fetched automatically and refreshed once a day)
- The token was issued **for your app** specifically (matches your `googleClientID`) — this stops someone using a Google token meant for a *different* app to log into yours

If both checks pass, you get back the user's `Email` and `EmailVerified` status — you can now create a session/JWT for them using `auth.GenerateJWT` from above.

---

## `database`

**What it is for:** Opening a connection to a **PostgreSQL** database, with sensible defaults already set (max/min connections, connection lifetime), so you don't set this up by hand in every project.

**How to use it:**

```go
import (
    "context"
    "github.com/dee25092005/go-core-pkg/database"
)

pool, err := database.NewPool(context.Background(), "postgres://user:pass@localhost:5432/mydb")
if err != nil {
    // handle error — e.g. wrong connection string, or database is unreachable
}
defer pool.Close()

// pool is a *pgxpool.Pool (from jackc/pgx/v5) — use it like normal
rows, err := pool.Query(context.Background(), "SELECT id, name FROM users")
```

`NewPool` already checks the connection works (it pings the database) before returning, so if `err` is `nil`, you know the pool is ready to use.

---

## `middleware`

Three Echo middlewares you plug into your server.

### 1. `AuthMiddleware` — checks the login token on every request

```go
import "github.com/dee25092005/go-core-pkg/middleware"

e := echo.New()

api := e.Group("/api")
api.Use(middleware.AuthMiddleware("my-secret-key"))
```

- Reads the `Authorization: Bearer <token>` header
- Verifies the token using `auth.VerifyJWT`
- If it's valid, saves `user_id` and `tier` into the request context so your handlers can read them:

```go
func handler(c echo.Context) error {
    userID := c.Get("user_id").(string)
    tier := c.Get("tier").(string)
    ...
}
```

- If it's missing or invalid, it stops the request and returns a 401 error automatically.

### 2. `CasbinAuthZ` — checks if the user is *allowed* to do something

While `AuthMiddleware` checks *who* you are (login), `CasbinAuthZ` checks *what you're allowed to do* (permissions) — using [Casbin](https://casbin.org/), a permission/access-control library. Run this one **after** `AuthMiddleware`, since it needs to know who the user is first.

```go
import (
    "github.com/casbin/casbin/v2"
    "github.com/dee25092005/go-core-pkg/middleware"
)

// set up the enforcer once, using your model + policy files
enforcer, err := casbin.NewEnforcer("model.conf", "policy.csv")
if err != nil {
    // handle error
}

api := e.Group("/api")
api.Use(middleware.AuthMiddleware("my-secret-key"))
api.Use(middleware.CasbinAuthZ(enforcer))
```

How it decides: it reads the `X-user` header (defaults to `"anonymous"` if missing), together with the request path and HTTP method (e.g. `GET /api/files`), and asks the `enforcer` "is this user allowed to do this?" based on your `model.conf`/`policy.csv` rules.

- If **allowed** → request continues normally.
- If **not allowed** → stops the request and returns a 403 Forbidden error automatically.

> **Note:** right now it reads the username from the `X-user` header rather than from the JWT claims that `AuthMiddleware` sets in context (`c.Get("user_id")`). If you want permission checks tied to the logged-in user from the token instead of a header, that's a small change worth making — let me know if you want it wired that way.

### 3. `CustomHTTPErrorHandler` — turns any error into clean JSON

```go
e := echo.New()
e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler
```

Set this once when you create your Echo app. From then on, whenever a handler returns an error (whether it's an `apperrors.AppError`, a normal Echo error, or an unexpected Go error), it gets converted into the same JSON error shape shown above — so your API is always consistent.

### 4. `RequestLogger` — logs every request

```go
e.Use(middleware.RequestLogger())
```

Logs each request's method, URL, status code, response size, how long it took, and the caller's IP. Useful for seeing what's happening on your server without adding print statements everywhere.

---

## `phajay`

**What it is for:** Talking to [PhaJay](https://phajay.co), a Lao payment gateway, to generate QR codes for bank payments (BCEL, JDB, LDB, IB, STB) and get notified when a payment is completed.

**How to use it:**

```go
import (
    "context"
    "github.com/dee25092005/go-core-pkg/phajay"
)

client := phajay.NewClient("my-phajay-secret-key")

resp, err := client.GenerateQR(context.Background(), "bcel", phajay.GenerateQRRequest{
    Amount:      50000,
    Description: "Order #1234",
})
if err != nil {
    // handle error
}

fmt.Println(resp.QRCode)        // the QR code to show the user
fmt.Println(resp.TransactionID) // save this to match the payment later
```

**Bank codes you can use:** `"bcel"`, `"jdb"`, `"ldb"`, `"ib"`, `"stb"`

**Getting notified when the payment is done** (real-time, via Socket.IO):

```go
socketClient := phajay.NewSokcetClient("my-phajay-secret-key")

socketClient.StartListener(context.Background(), func(txID string, data map[string]interface{}) {
    // this runs automatically when a payment completes
    fmt.Println("Payment completed for transaction:", txID)
    // e.g. mark the order as paid in your database
})
```

---

## `response`

**What it is for:** One simple shape for successful responses, so every "OK" reply from your API looks the same.

**How to use it:**

```go
import "github.com/dee25092005/go-core-pkg/response"

return c.JSON(200, response.SuccessResponse{
    Data:    user,
    Message: "user created",
})
```

Output:

```json
{
  "data": { ...user... },
  "message": "user created"
}
```

---

## `storage`

**What it is for:** Uploading and deleting files on **Cloudflare R2** (which works the same way as AWS S3).

**How to use it:**

```go
import (
    "context"
    "github.com/dee25092005/go-core-pkg/storage"
)

store, err := storage.NewR2Storage(context.Background(), storage.R2Config{
    AccountID:       "your-cloudflare-account-id",
    AccessKeyID:     "your-r2-access-key",
    AccessKeySecret: "your-r2-secret-key",
    BucketName:      "my-bucket",
    PublicURL:       "https://files.example.com",
})
if err != nil {
    // handle error
}

// Upload a file
url, err := store.UploadFile(context.Background(), "avatars/user1.png", fileReader, "image/png")
// url = "https://files.example.com/avatars/user1.png"

// Delete a file
err = store.DeleteFile(context.Background(), "avatars/user1.png")
```

---

## `utils`

Small helper functions.

```go
import "github.com/dee25092005/go-core-pkg/utils"

// Make a random token (good for things like password-reset links)
token, err := utils.GenerateRandomToken()

// Hash a token before saving it to your database
// (so if your DB leaks, the real token is still safe)
hashed := utils.HashToken(token)
```

---

## Quick start example

Putting it all together in a small Echo server:

```go
package main

import (
    "github.com/dee25092005/go-core-pkg/middleware"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()

    e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler
    e.Use(middleware.RequestLogger())

    api := e.Group("/api")
    api.Use(middleware.AuthMiddleware("my-secret-key"))

    api.GET("/me", func(c echo.Context) error {
        userID := c.Get("user_id").(string)
        return c.JSON(200, map[string]string{"user_id": userID})
    })

    e.Logger.Fatal(e.Start(":8080"))
}
```

---

## Setup notes

- Requires **Go 1.26.5+**
- After pulling changes, run:
  ```
  go mod tidy
  ```
  to make sure all dependencies (like `golang-jwt/jwt/v5`) are downloaded correctly.
