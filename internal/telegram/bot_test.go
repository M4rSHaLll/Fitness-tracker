package telegram

import (
	"Fitness-tracker/internal/repository/memory"
	"Fitness-tracker/internal/service"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestSanitizeTelegramError(t *testing.T) {
	const token = "123456:secret"
	err := errors.New("Post \"https://api.telegram.org/bot" + token + "/getMe\": timeout")

	message := sanitizeTelegramError(err, token)

	if strings.Contains(message, token) {
		t.Fatal("sanitized error contains bot token")
	}
	if !strings.Contains(message, "[redacted]") {
		t.Fatalf("expected redaction marker, got %q", message)
	}
}

func TestBotProfileExerciseAndWorkoutFlow(t *testing.T) {
	bot := newTestBot()

	if _, err := bot.createExerciseText("Bench press"); err != nil {
		t.Fatalf("create exercise: %v", err)
	}
	if _, err := bot.updateProfileText(1001, "art", "80 180 30"); err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if _, err := bot.createWorkoutText(1001, "art", "Push day"); err != nil {
		t.Fatalf("create workout: %v", err)
	}

	workouts, err := bot.workoutsText(1001, "art")
	if err != nil {
		t.Fatalf("get workouts text: %v", err)
	}
	if !strings.Contains(workouts, "Push day") {
		t.Fatalf("expected workout description, got %q", workouts)
	}

	user, err := bot.services.Users.GetUserByTelegramID(1001)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.Weight != 80 || user.Height != 180 || user.Age != 30 {
		t.Fatalf("unexpected profile: %+v", user)
	}
}

func TestInteractiveSetFlow(t *testing.T) {
	bot := newTestBot()
	exercise, err := bot.services.Exercises.CreateExercise("Squat")
	if err != nil {
		t.Fatalf("create exercise: %v", err)
	}
	user, err := bot.services.Users.GetOrCreateTelegramUser(2001, "lifter")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	workout, err := bot.services.Workouts.CreateWorkout(user.ID, "Leg day")
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}

	telegramUser := &tgbotapi.User{ID: 2001, UserName: "lifter"}
	text, keyboard, err := bot.callbackResponse(&tgbotapi.CallbackQuery{
		From: telegramUser,
		Data: callbackWorkoutPrefix + strconv.FormatInt(workout.ID, 10),
	})
	if err != nil {
		t.Fatalf("select workout: %v", err)
	}
	if text != "Выберите упражнение:" || len(keyboard.InlineKeyboard) != 1 {
		t.Fatalf("unexpected workout callback response: %q, rows=%d", text, len(keyboard.InlineKeyboard))
	}

	_, _, err = bot.callbackResponse(&tgbotapi.CallbackQuery{
		From: telegramUser,
		Data: callbackExercisePrefix + strconv.FormatInt(exercise.ID, 10),
	})
	if err != nil {
		t.Fatalf("select exercise: %v", err)
	}

	handled, text, err := bot.handleSetValues(&tgbotapi.Message{
		From: telegramUser,
		Text: "100 5 8.5",
	})
	if err != nil {
		t.Fatalf("create set: %v", err)
	}
	if !handled || !strings.Contains(text, "Подход создан") {
		t.Fatalf("unexpected set response: handled=%v text=%q", handled, text)
	}

	sets, err := bot.services.Sets.GetWorkoutSets(workout.ID)
	if err != nil {
		t.Fatalf("get workout sets: %v", err)
	}
	if len(sets) != 1 || sets[0].Weight != 100 || sets[0].Reps != 5 || sets[0].RPE != 8.5 {
		t.Fatalf("unexpected sets: %+v", sets)
	}
	if _, ok := bot.getDraft(telegramUser.ID); ok {
		t.Fatal("expected completed draft to be removed")
	}
}

func TestInteractiveSetRejectsForeignWorkout(t *testing.T) {
	bot := newTestBot()
	owner, err := bot.services.Users.GetOrCreateTelegramUser(3001, "owner")
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	workout, err := bot.services.Workouts.CreateWorkout(owner.ID, "Owner workout")
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}

	text, _, err := bot.callbackResponse(&tgbotapi.CallbackQuery{
		From: &tgbotapi.User{ID: 3002, UserName: "foreign"},
		Data: callbackWorkoutPrefix + strconv.FormatInt(workout.ID, 10),
	})
	if err != nil {
		t.Fatalf("select foreign workout: %v", err)
	}
	if !strings.Contains(text, "другому пользователю") {
		t.Fatalf("expected ownership error, got %q", text)
	}
	if _, ok := bot.getDraft(3002); ok {
		t.Fatal("foreign user draft must not be created")
	}
}

func TestParseSetValues(t *testing.T) {
	weight, reps, rpe, err := parseSetValues("102.5 6 9.5")
	if err != nil {
		t.Fatalf("parse values: %v", err)
	}
	if weight != 102.5 || reps != 6 || rpe != 9.5 {
		t.Fatalf("unexpected values: %v %d %v", weight, reps, rpe)
	}

	if _, _, _, err := parseSetValues("100 five 8"); err == nil {
		t.Fatal("expected invalid reps error")
	}
}

func newTestBot() *Bot {
	userRepo := memory.NewUserRepository()
	workoutRepo := memory.NewWorkoutRepository()
	setRepo := memory.NewSetRepository()
	exerciseRepo := memory.NewExerciseRepository()

	return &Bot{
		services: Services{
			Users:     service.NewUserService(userRepo),
			Workouts:  service.NewWorkoutService(workoutRepo, userRepo),
			Sets:      service.NewSetService(setRepo, workoutRepo, exerciseRepo),
			Stats:     service.NewStatsService(setRepo, workoutRepo, userRepo),
			Exercises: service.NewExerciseService(exerciseRepo),
		},
		logger: log.New(io.Discard, "", 0),
		drafts: make(map[int64]setDraft),
	}
}
