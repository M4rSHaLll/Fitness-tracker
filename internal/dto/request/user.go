package request

type CreateUser struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
}

type UpdateUser struct {
	Weight float64 `json:"weight"`
	Height float64 `json:"height"`
	Age    int64   `json:"age"`
}
