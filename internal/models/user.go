
package models

import (
    "context"
    "errors"
    "golang.org/x/crypto/bcrypt"
    "github.com/rajnish-012/delivery-management-system/internal/database"
)

type User struct {
    ID           int
    Username     string
    PasswordHash string
    Role         string // "customer" or "admin"
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
    return err == nil
}

func CreateUser(ctx context.Context, username, password, role string) (*User, error) {
    if role != "customer" && role != "admin" {
        return nil, errors.New("invalid role")
    }
    pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    var id int
    err = database.Pool.QueryRow(ctx,
        "INSERT INTO users (username, password_hash, role) VALUES ($1,$2,$3) RETURNING id",
        username, string(pwHash), role,
    ).Scan(&id)
    if err != nil {
        return nil, err
    }
    return &User{ID: id, Username: username, PasswordHash: string(pwHash), Role: role}, nil
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
    u := &User{}
    row := database.Pool.QueryRow(ctx, "SELECT id, username, password_hash, role FROM users WHERE username=$1", username)
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role); err != nil {
        return nil, err
    }
    return u, nil
}

func GetUserByID(ctx context.Context, id int) (*User, error) {
    u := &User{}
    row := database.Pool.QueryRow(ctx, "SELECT id, username, password_hash, role FROM users WHERE id=$1", id)
    if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role); err != nil {
        return nil, err
    }
    return u, nil
}
