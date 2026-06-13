package response

import "time"

type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Weight float64 `json:"weight"`
	Height float64 `json:"height"`
	Age    int64   `json:"age"`
}
