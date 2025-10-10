package auth

import (
    "context"
    "errors"
    "os"
    "strconv"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "net/http"
)

var jwtSecret = []byte(func() string {
    s := os.Getenv("JWT_SECRET")
    if s == "" {
        s = "dev-secret" // change in production
    }
    return s
}())

type Claims struct {
    UserID int    `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int, role string) (string, error) {
    expMinutes := 60
    if v := os.Getenv("JWT_EXP_MINUTES"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            expMinutes = n
        }
    }
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expMinutes) * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}

// Middleware for routes (simple)
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth == "" {
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }
        // Expect "Bearer <token>"
        var tokenStr string
        if len(auth) > 7 && auth[:7] == "Bearer " {
            tokenStr = auth[7:]
        } else {
            http.Error(w, "invalid authorization header", http.StatusUnauthorized)
            return
        }
        claims, err := ParseToken(tokenStr)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        // attach to context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
