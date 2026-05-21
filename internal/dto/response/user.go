package response

type User struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`

	Weight float64 `json:"weight"`
	Height float64 `json:"height"`
	Age    int64   `json:"age"`
}
