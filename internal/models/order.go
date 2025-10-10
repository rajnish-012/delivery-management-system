package models

import (
    "context"
    "time"

    "github.com/rajnish-012/delivery-management-system/internal/database"
)

type Order struct {
    ID         int       `json:"id"`
    CustomerID int       `json:"customer_id"`
    Item       string    `json:"item"`
    Status     string    `json:"status"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func CreateOrder(ctx context.Context, customerID int, item string) (*Order, error) {
    var id int
    err := database.Pool.QueryRow(ctx,
        "INSERT INTO orders (customer_id, item, status) VALUES ($1,$2,$3) RETURNING id",
        customerID, item, "created",
    ).Scan(&id)
    if err != nil {
        return nil, err
    }
    return GetOrderByID(ctx, id)
}

func GetOrderByID(ctx context.Context, id int) (*Order, error) {
    o := &Order{}
    row := database.Pool.QueryRow(ctx, "SELECT id, customer_id, item, status, created_at, updated_at FROM orders WHERE id=$1", id)
    if err := row.Scan(&o.ID, &o.CustomerID, &o.Item, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
        return nil, err
    }
    return o, nil
}

func UpdateOrderStatus(ctx context.Context, id int, status string) error {
    _, err := database.Pool.Exec(ctx, "UPDATE orders SET status=$1, updated_at=now() WHERE id=$2", status, id)
    return err
}

func CancelOrder(ctx context.Context, id int) error {
    // only set cancelled if not delivered
    _, err := database.Pool.Exec(ctx, "UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1 AND status != 'delivered'", id)
    return err
}

// List orders (admin/all or by customer)
func ListOrdersByCustomer(ctx context.Context, customerID int) ([]*Order, error) {
    rows, err := database.Pool.Query(ctx, "SELECT id, customer_id, item, status, created_at, updated_at FROM orders WHERE customer_id=$1 ORDER BY created_at DESC", customerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var res []*Order
    for rows.Next() {
        o := &Order{}
        if err := rows.Scan(&o.ID, &o.CustomerID, &o.Item, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
            return nil, err
        }
        res = append(res, o)
    }
    return res, nil
}

func ListAllOrders(ctx context.Context) ([]*Order, error) {
    rows, err := database.Pool.Query(ctx, "SELECT id, customer_id, item, status, created_at, updated_at FROM orders ORDER BY created_at DESC")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var res []*Order
    for rows.Next() {
        o := &Order{}
        if err := rows.Scan(&o.ID, &o.CustomerID, &o.Item, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
            return nil, err
        }
        res = append(res, o)
    }
    return res, nil
}
