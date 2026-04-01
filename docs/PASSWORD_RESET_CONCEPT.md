# Password Reset Concept: Secure Token-Based Flow

This document outlines the architectural and security design for the password reset mechanism in Cogni-Cash.

## 1. Objective
To provide a secure, user-friendly way for users to regain access to their accounts if they forget their password, utilizing the established SMTP infrastructure.

## 2. The Flow (End-to-End)

### Phase A: Request
1. **User Action:** The user clicks "Forgot Password" on the Login page and enters their registered email address.
2. **Backend Validation:**
    - The system checks if a user with that email exists.
    - If **not found**, the system returns a generic "Success" message to prevent email enumeration attacks.
    - If **found**:
        - Generate a cryptographically secure random token (e.g., 32-byte hex string).
        - Hash the token (SHA-256) for storage.
        - Create a record in `password_reset_tokens` linked to the user ID, storing the `token_hash` and an `expires_at` timestamp (default: 1 hour).
        - Invalidate any previous active reset tokens for this user.

### Phase B: Notification
1. **Email Delivery:** The `NotificationService` sends an email containing a unique link:
   `https://<domain>/reset-password?token=<plain_token>`
2. **Security:** The plain token is *never* stored in the database, only its hash.

### Phase C: Reset
1. **User Action:** User clicks the link and is navigated to a "New Password" form in the React frontend.
2. **Token Verification:**
    - The frontend sends the plain token to the backend for validation.
    - The backend hashes the provided token and looks up the record.
    - **Validation Criteria:** Token must exist, match the hash, and `expires_at > NOW()`.
3. **Password Update:**
    - If valid, the user submits a new password.
    - The backend hashes the new password (bcrypt) and updates the `users` table.
    - The reset token record is **immediately deleted** to prevent reuse.
    - (Optional) A confirmation email "Your password was changed" is sent.

## 3. Security Requirements

### Token Generation
- **Source:** Must use a cryptographically secure pseudo-random number generator (CSPRNG), specifically `crypto/rand` in Go.
- **Entropy:** Minimum 256 bits of entropy.

### Storage (At Rest)
- **Hash before store:** We store `SHA-256(token)`. If the database is compromised, an attacker cannot use the stolen hashes to reset passwords because they cannot reverse the hash to get the plain token needed for the URL.

### Timing Attacks
- The verification logic should ideally be constant-time when comparing hashes to prevent side-channel leaks, though standard SQL lookups are generally acceptable for this use case given the high entropy of the tokens.

### Rate Limiting
- **Request Limit:** Limit the number of reset requests per email/IP (e.g., 3 requests per hour) to prevent SMTP spamming.
- **Attempt Limit:** Limit the number of failed token verification attempts.

## 4. Database Schema

### `password_reset_tokens`
| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK` | Internal identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | The user requesting the reset |
| `token_hash` | `TEXT` | `NOT NULL`, `UNIQUE` | SHA-256 hash of the random token |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` | Expiration (usually +1h) |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT NOW()` | Audit timestamp |

## 5. API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/forgot-password` | Public | Initiates the process, sends email. |
| `GET` | `/api/v1/auth/reset-password/validate` | Public | Validates if a token is still active/valid. |
| `POST` | `/api/v1/auth/reset-password/confirm` | Public | Sets the new password and consumes the token. |
