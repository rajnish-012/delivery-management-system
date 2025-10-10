package tests

import (
    "context"
    "testing"
    "github.com/rajnish-012/delivery-management-system/internal/auth"
    "github.com/rajnish-012/delivery-management-system/internal/database"
    "github.com/rajnish-012/delivery-management-system/internal/models"
)

func TestUserAndOrderLifecycle(t *testing.T) {
    ctx := context.Background()
    if err := database.InitPostgres(ctx); err != nil {
        t.Fatalf("pg init: %v", err)
    }
    defer database.ClosePostgres()
    if err := database.InitRedis(ctx); err != nil {
        t.Fatalf("redis init: %v", err)
    }

    // cleanup tables (safe for tests)
    database.Pool.Exec(ctx, "TRUNCATE orders, users RESTART IDENTITY CASCADE")

    // create user
    u, err := models.CreateUser(ctx, "testuser", "pass123", "customer")
    if err != nil {
        t.Fatalf("create user: %v", err)
    }

    // generate token (basic sanity)
    if _, err := auth.GenerateToken(u.ID, u.Role); err != nil {
        t.Fatalf("jwt: %v", err)
    }

    // create order
    ord, err := models.CreateOrder(ctx, u.ID, "book")
    if err != nil {
        t.Fatalf("create order: %v", err)
    }

    // Start progression
    // use smaller waits by temporarily overriding lifecycle wait in production code you'd parameterize this
    // For now just wait a bit and check status changes
    // start background progression
    // (use orders.StartProgression)
    // to avoid import cycle, run minimal check: Update status then verify Cancel works

    if err := models.UpdateOrderStatus(ctx, ord.ID, "dispatched"); err != nil {
        t.Fatalf("update: %v", err)
    }

    o2, err := models.GetOrderByID(ctx, ord.ID)
    if err != nil {
        t.Fatalf("get order: %v", err)
    }
    if o2.Status != "dispatched" {
        t.Fatalf("expected dispatched, got %s", o2.Status)
    }

    // cancel
    if err := models.CancelOrder(ctx, ord.ID); err != nil {
        t.Fatalf("cancel: %v", err)
    }
    o3, _ := models.GetOrderByID(ctx, ord.ID)
    if o3.Status != "cancelled" {
        t.Fatalf("expected cancelled got %s", o3.Status)
    }

    // done
}
