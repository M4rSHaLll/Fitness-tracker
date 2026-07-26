package telegram

import (
	"Fitness-tracker/internal/service"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	initialRequestTimeout = 15 * time.Second
	pollingRequestTimeout = 35 * time.Second
	pollingTimeoutSeconds = 25
	pollingRetryInterval  = 3 * time.Second
	maxWorkoutButtons     = 10
	maxExerciseButtons    = 30
)

const (
	callbackWorkoutPrefix  = "set:workout:"
	callbackExercisePrefix = "set:exercise:"
)

type Services struct {
	Users     *service.UserService
	Workouts  *service.WorkoutService
	Sets      *service.SetService
	Stats     *service.StatsService
	Exercises *service.ExerciseService
}

type Bot struct {
	api      *tgbotapi.BotAPI
	services Services
	logger   *log.Logger
	drafts   map[int64]setDraft
	draftsMu sync.RWMutex
}

type setDraft struct {
	WorkoutID  int64
	ExerciseID int64
}

func NewBot(token string, services Services, logger *log.Logger) (*Bot, error) {
	client := &http.Client{Timeout: initialRequestTimeout}
	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		return nil, fmt.Errorf("connect to Telegram API: %s", sanitizeTelegramError(err, token))
	}
	api.Client = &http.Client{Timeout: pollingRequestTimeout}

	return &Bot{
		api:      api,
		services: services,
		logger:   logger,
		drafts:   make(map[int64]setDraft),
	}, nil
}

func sanitizeTelegramError(err error, token string) string {
	message := err.Error()
	if token == "" {
		return message
	}
	return strings.ReplaceAll(message, token, "[redacted]")
}

func (b *Bot) Start(ctx context.Context) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = pollingTimeoutSeconds
	b.logger.Printf("telegram bot started as @%s", b.api.Self.UserName)

	for {
		updates, err := b.getUpdates(ctx, updateConfig)
		if err != nil {
			b.logTelegramError("get Telegram updates", err)
			if !waitForContext(ctx, pollingRetryInterval) {
				b.logger.Print("telegram bot stopped")
				return
			}
			continue
		}
		if ctx.Err() != nil {
			b.logger.Print("telegram bot stopped")
			return
		}

		for _, update := range updates {
			if update.UpdateID >= updateConfig.Offset {
				updateConfig.Offset = update.UpdateID + 1
			}
			b.handleUpdate(update)
		}
	}
}

func (b *Bot) getUpdates(ctx context.Context, config tgbotapi.UpdateConfig) ([]tgbotapi.Update, error) {
	type result struct {
		updates []tgbotapi.Update
		err     error
	}

	resultCh := make(chan result, 1)
	go func() {
		updates, err := b.api.GetUpdates(config)
		resultCh <- result{updates: updates, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		return result.updates, result.err
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
		return
	}
	if update.Message == nil || update.Message.From == nil {
		return
	}
	if !update.Message.IsCommand() {
		handled, text, err := b.handleSetValues(update.Message)
		if handled {
			if err != nil {
				text = userFacingError(err)
			}
			b.reply(update.Message.Chat.ID, text)
			return
		}
		b.reply(update.Message.Chat.ID, helpText())
		return
	}

	text, err := b.handleCommand(update.Message)
	if err != nil {
		text = userFacingError(err)
	}
	if text != "" {
		b.reply(update.Message.Chat.ID, text)
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) (string, error) {
	telegramID := message.From.ID
	username := telegramUsername(message.From)

	switch message.Command() {
	case "start":
		user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Готово, профиль найден или создан. Ваш внутренний user_id: %d\n\n%s", user.ID, helpText()), nil
	case "help":
		return helpText(), nil
	case "profile":
		user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Профиль: %s\nTelegram ID: %d\nВес: %.1f\nРост: %.1f\nВозраст: %d", user.Username, user.TelegramID, user.Weight, user.Height, user.Age), nil
	case "profile_set":
		return b.updateProfileText(telegramID, username, message.CommandArguments())
	case "exercises":
		return b.exercisesText()
	case "exercise":
		return b.createExerciseText(message.CommandArguments())
	case "workout":
		return b.createWorkoutText(telegramID, username, message.CommandArguments())
	case "workouts":
		return b.workoutsText(telegramID, username)
	case "set":
		return b.createSetText(telegramID, username, message.CommandArguments())
	case "addset":
		text, keyboard, err := b.startSetFlow(telegramID, username)
		if err != nil {
			return "", err
		}
		b.replyWithMarkup(message.Chat.ID, text, keyboard)
		return "", nil
	case "cancel":
		b.deleteDraft(telegramID)
		return "Текущее действие отменено.", nil
	case "stats":
		return b.statsText(telegramID, username)
	default:
		return helpText(), nil
	}
}

func (b *Bot) exercisesText() (string, error) {
	exercises, err := b.services.Exercises.GetExercises()
	if err != nil {
		return "", err
	}
	if len(exercises) == 0 {
		return "Упражнений пока нет. Добавьте их через API: POST /exercises.", nil
	}

	var builder strings.Builder
	builder.WriteString("Упражнения:\n")
	for _, exercise := range exercises {
		builder.WriteString(fmt.Sprintf("%d. %s\n", exercise.ID, exercise.Name))
	}

	return strings.TrimSpace(builder.String()), nil
}

func (b *Bot) createExerciseText(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Использование: /exercise Жим лежа", nil
	}

	exercise, err := b.services.Exercises.CreateExercise(name)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Упражнение добавлено: %s", exercise.Name), nil
}

func (b *Bot) updateProfileText(telegramID int64, username, args string) (string, error) {
	fields := strings.Fields(args)
	if len(fields) != 3 {
		return "Использование: /profile_set <вес> <рост> <возраст>\nПример: /profile_set 80 180 30", nil
	}

	weight, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "Вес должен быть числом.", nil
	}
	height, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return "Рост должен быть числом.", nil
	}
	age, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return "Возраст должен быть целым числом.", nil
	}

	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", err
	}
	if err := b.services.Users.UpdateProfile(user.ID, weight, height, age); err != nil {
		return "", err
	}

	return fmt.Sprintf("Профиль обновлен: %.1f кг, %.1f см, %d лет.", weight, height, age), nil
}

func (b *Bot) createWorkoutText(telegramID int64, username, description string) (string, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return "Использование: /workout Push day", nil
	}

	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", err
	}

	workout, err := b.services.Workouts.CreateWorkout(user.ID, description)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Тренировка создана: #%d\n%s", workout.ID, workout.Description), nil
}

func (b *Bot) workoutsText(telegramID int64, username string) (string, error) {
	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", err
	}

	workouts, err := b.services.Workouts.GetUserWorkouts(user.ID)
	if err != nil {
		return "", err
	}
	if len(workouts) == 0 {
		return "Тренировок пока нет. Создайте первую: /workout Push day", nil
	}

	var builder strings.Builder
	builder.WriteString("Последние тренировки:\n")
	start := max(0, len(workouts)-maxWorkoutButtons)
	for i := len(workouts) - 1; i >= start; i-- {
		workout := workouts[i]
		builder.WriteString(fmt.Sprintf("#%d — %s\n", workout.ID, workout.Description))
	}

	return strings.TrimSpace(builder.String()), nil
}

func (b *Bot) createSetText(telegramID int64, username, args string) (string, error) {
	fields := strings.Fields(args)
	if len(fields) != 5 {
		return "Использование: /set <workout_id> <exercise_id> <weight> <reps> <rpe>\nПример: /set 1 2 100 5 8.5", nil
	}

	workoutID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return "workout_id должен быть числом", nil
	}
	exerciseID, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return "exercise_id должен быть числом", nil
	}
	weight, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return "weight должен быть числом", nil
	}
	reps, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return "reps должен быть числом", nil
	}
	rpe, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return "rpe должен быть числом", nil
	}

	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", err
	}

	workout, err := b.services.Workouts.GetWorkoutByID(workoutID)
	if err != nil {
		return "", err
	}
	if workout.UserID != user.ID {
		return "Эта тренировка принадлежит другому пользователю.", nil
	}

	set, err := b.services.Sets.CreateSet(exerciseID, workoutID, reps, weight, rpe)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Подход создан: #%d\n%d повторений x %.1f кг, RPE %.1f", set.ID, set.Reps, set.Weight, set.RPE), nil
}

func (b *Bot) startSetFlow(telegramID int64, username string) (string, tgbotapi.InlineKeyboardMarkup, error) {
	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, err
	}

	workouts, err := b.services.Workouts.GetUserWorkouts(user.ID)
	if err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, err
	}
	if len(workouts) == 0 {
		return "Сначала создайте тренировку: /workout Push day", tgbotapi.InlineKeyboardMarkup{}, nil
	}

	start := max(0, len(workouts)-maxWorkoutButtons)
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(workouts)-start)
	for i := len(workouts) - 1; i >= start; i-- {
		workout := workouts[i]
		label := fmt.Sprintf("#%d %s", workout.ID, truncateText(workout.Description, 40))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, callbackWorkoutPrefix+strconv.FormatInt(workout.ID, 10)),
		))
	}

	return "Выберите тренировку:", tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	if callback.From == nil || callback.Message == nil || callback.Message.Chat == nil {
		return
	}
	b.answerCallback(callback.ID)

	text, keyboard, err := b.callbackResponse(callback)
	if err != nil {
		text = userFacingError(err)
	}
	b.replyWithMarkup(callback.Message.Chat.ID, text, keyboard)
}

func (b *Bot) callbackResponse(callback *tgbotapi.CallbackQuery) (string, tgbotapi.InlineKeyboardMarkup, error) {
	telegramID := callback.From.ID
	username := telegramUsername(callback.From)

	switch {
	case strings.HasPrefix(callback.Data, callbackWorkoutPrefix):
		workoutID, err := parseCallbackID(callback.Data, callbackWorkoutPrefix)
		if err != nil {
			return "Кнопка устарела. Запустите /addset еще раз.", tgbotapi.InlineKeyboardMarkup{}, nil
		}

		user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
		if err != nil {
			return "", tgbotapi.InlineKeyboardMarkup{}, err
		}
		workout, err := b.services.Workouts.GetWorkoutByID(workoutID)
		if err != nil {
			return "", tgbotapi.InlineKeyboardMarkup{}, err
		}
		if workout.UserID != user.ID {
			return "Эта тренировка принадлежит другому пользователю.", tgbotapi.InlineKeyboardMarkup{}, nil
		}

		exercises, err := b.services.Exercises.GetExercises()
		if err != nil {
			return "", tgbotapi.InlineKeyboardMarkup{}, err
		}
		if len(exercises) == 0 {
			return "Сначала добавьте упражнение: /exercise Жим лежа", tgbotapi.InlineKeyboardMarkup{}, nil
		}

		b.setDraft(telegramID, setDraft{WorkoutID: workoutID})
		limit := min(len(exercises), maxExerciseButtons)
		rows := make([][]tgbotapi.InlineKeyboardButton, 0, limit)
		for _, exercise := range exercises[:limit] {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					truncateText(exercise.Name, 48),
					callbackExercisePrefix+strconv.FormatInt(exercise.ID, 10),
				),
			))
		}
		return "Выберите упражнение:", tgbotapi.NewInlineKeyboardMarkup(rows...), nil

	case strings.HasPrefix(callback.Data, callbackExercisePrefix):
		exerciseID, err := parseCallbackID(callback.Data, callbackExercisePrefix)
		if err != nil {
			return "Кнопка устарела. Запустите /addset еще раз.", tgbotapi.InlineKeyboardMarkup{}, nil
		}

		draft, ok := b.getDraft(telegramID)
		if !ok || draft.WorkoutID == 0 {
			return "Сессия добавления подхода не найдена. Запустите /addset.", tgbotapi.InlineKeyboardMarkup{}, nil
		}
		if _, err := b.services.Exercises.GetExerciseByID(exerciseID); err != nil {
			return "", tgbotapi.InlineKeyboardMarkup{}, err
		}

		draft.ExerciseID = exerciseID
		b.setDraft(telegramID, draft)
		return "Введите вес, повторения и RPE через пробел.\nПример: 100 5 8.5\nДля отмены: /cancel", tgbotapi.InlineKeyboardMarkup{}, nil
	default:
		return "Кнопка устарела. Запустите /addset еще раз.", tgbotapi.InlineKeyboardMarkup{}, nil
	}
}

func (b *Bot) handleSetValues(message *tgbotapi.Message) (bool, string, error) {
	draft, ok := b.getDraft(message.From.ID)
	if !ok || draft.WorkoutID == 0 || draft.ExerciseID == 0 {
		return false, "", nil
	}

	weight, reps, rpe, err := parseSetValues(message.Text)
	if err != nil {
		return true, "Введите три значения: вес, повторения и RPE.\nПример: 100 5 8.5\nДля отмены: /cancel", nil
	}

	user, err := b.services.Users.GetOrCreateTelegramUser(message.From.ID, telegramUsername(message.From))
	if err != nil {
		return true, "", err
	}
	workout, err := b.services.Workouts.GetWorkoutByID(draft.WorkoutID)
	if err != nil {
		b.deleteDraft(message.From.ID)
		return true, "", err
	}
	if workout.UserID != user.ID {
		b.deleteDraft(message.From.ID)
		return true, "Эта тренировка принадлежит другому пользователю.", nil
	}

	set, err := b.services.Sets.CreateSet(draft.ExerciseID, draft.WorkoutID, reps, weight, rpe)
	if err != nil {
		return true, "", err
	}
	b.deleteDraft(message.From.ID)

	return true, fmt.Sprintf("Подход создан: #%d\n%d повторений x %.1f кг, RPE %.1f", set.ID, set.Reps, set.Weight, set.RPE), nil
}

func parseSetValues(text string) (float64, int64, float64, error) {
	fields := strings.Fields(text)
	if len(fields) != 3 {
		return 0, 0, 0, errors.New("expected weight, reps and rpe")
	}

	weight, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	reps, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	rpe, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return weight, reps, rpe, nil
}

func parseCallbackID(data, prefix string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(data, prefix), 10, 64)
}

func (b *Bot) setDraft(telegramID int64, draft setDraft) {
	b.draftsMu.Lock()
	defer b.draftsMu.Unlock()
	b.drafts[telegramID] = draft
}

func (b *Bot) getDraft(telegramID int64) (setDraft, bool) {
	b.draftsMu.RLock()
	defer b.draftsMu.RUnlock()
	draft, ok := b.drafts[telegramID]
	return draft, ok
}

func (b *Bot) deleteDraft(telegramID int64) {
	b.draftsMu.Lock()
	defer b.draftsMu.Unlock()
	delete(b.drafts, telegramID)
}

func (b *Bot) statsText(telegramID int64, username string) (string, error) {
	user, err := b.services.Users.GetOrCreateTelegramUser(telegramID, username)
	if err != nil {
		return "", err
	}

	stats, err := b.services.Stats.GetUserStats(user.ID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Статистика:\nТренировок: %d\nОбъем: %.1f\nСредний RPE: %.1f", stats.TotalWorkouts, stats.TotalVolume, stats.AverageRPE), nil
}

func (b *Bot) reply(chatID int64, text string) {
	b.replyWithMarkup(chatID, text, tgbotapi.InlineKeyboardMarkup{})
}

func (b *Bot) replyWithMarkup(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	message := tgbotapi.NewMessage(chatID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		message.ReplyMarkup = keyboard
	}
	if _, err := b.api.Send(message); err != nil {
		b.logTelegramError("send Telegram message", err)
	}
}

func (b *Bot) answerCallback(callbackID string) {
	if _, err := b.api.Request(tgbotapi.NewCallback(callbackID, "")); err != nil {
		b.logTelegramError("answer Telegram callback", err)
	}
}

func (b *Bot) logTelegramError(operation string, err error) {
	b.logger.Printf("%s: %s", operation, sanitizeTelegramError(err, b.api.Token))
}

func telegramUsername(user *tgbotapi.User) string {
	if user.UserName != "" {
		return user.UserName
	}
	name := strings.TrimSpace(strings.TrimSpace(user.FirstName + " " + user.LastName))
	if name != "" {
		return name
	}
	return fmt.Sprintf("telegram_%d", user.ID)
}

func userFacingError(err error) string {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return "Данные не прошли проверку. Проверьте формат команды и значения."
	case errors.Is(err, service.ErrUserNotFound):
		return "Пользователь не найден. Отправьте /start."
	case errors.Is(err, service.ErrWorkoutNotFound):
		return "Тренировка не найдена."
	case errors.Is(err, service.ErrExerciseNotFound):
		return "Упражнение не найдено. Проверьте /exercises."
	case errors.Is(err, service.ErrAlreadyExists):
		return "Такая запись уже существует."
	default:
		return "Не получилось выполнить команду. Попробуйте позже."
	}
}

func truncateText(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes-1]) + "…"
}

func waitForContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func helpText() string {
	return strings.TrimSpace(`Команды:
/start - создать или найти профиль
/profile - показать профиль
/profile_set <вес> <рост> <возраст> - обновить профиль
/exercises - список упражнений
/exercise <название> - добавить упражнение
/workout <описание> - создать тренировку
/workouts - последние тренировки
/addset - добавить подход с помощью кнопок
/set <workout_id> <exercise_id> <weight> <reps> <rpe> - добавить подход
/cancel - отменить добавление подхода
/stats - статистика`)
}
