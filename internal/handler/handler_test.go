package handler_test

import (
	"Fitness-tracker/internal/handler"
	"Fitness-tracker/internal/repository/memory"
	"Fitness-tracker/internal/router"
	"Fitness-tracker/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	server := newTestServer()

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"storage":"memory"`) {
		t.Fatalf("expected memory storage in response, got %s", resp.Body.String())
	}
}

func TestCreateUserRejectsDuplicateTelegramID(t *testing.T) {
	server := newTestServer()
	body := `{"telegram_id":123,"username":"art"}`

	first := doJSON(server, http.MethodPost, "/users", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d: %s", http.StatusCreated, first.Code, first.Body.String())
	}

	second := doJSON(server, http.MethodPost, "/users", body)
	if second.Code != http.StatusConflict {
		t.Fatalf("expected duplicate status %d, got %d: %s", http.StatusConflict, second.Code, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"code":"already_exists"`) {
		t.Fatalf("expected already_exists code, got %s", second.Body.String())
	}
}

func TestAPIRoutesRequireBearerToken(t *testing.T) {
	server := newTestServer()

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	server.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status %d, got %d: %s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"code":"invalid_api_token"`) {
		t.Fatalf("expected invalid_api_token code, got %s", resp.Body.String())
	}
}

func TestCreateWorkoutAndSetFlow(t *testing.T) {
	server := newTestServer()

	userResp := doJSON(server, http.MethodPost, "/users", `{"telegram_id":456,"username":"flow"}`)
	if userResp.Code != http.StatusCreated {
		t.Fatalf("create user status %d: %s", userResp.Code, userResp.Body.String())
	}
	userID := decodeID(t, userResp.Body.Bytes())

	exerciseResp := doJSON(server, http.MethodPost, "/exercises", `{"name":"Bench press"}`)
	if exerciseResp.Code != http.StatusCreated {
		t.Fatalf("create exercise status %d: %s", exerciseResp.Code, exerciseResp.Body.String())
	}
	exerciseID := decodeID(t, exerciseResp.Body.Bytes())

	workoutResp := doJSON(server, http.MethodPost, "/workouts", `{"user_id":`+itoa(userID)+`,"description":"Push"}`)
	if workoutResp.Code != http.StatusCreated {
		t.Fatalf("create workout status %d: %s", workoutResp.Code, workoutResp.Body.String())
	}
	workoutID := decodeID(t, workoutResp.Body.Bytes())

	setResp := doJSON(server, http.MethodPost, "/sets", `{"workout_id":`+itoa(workoutID)+`,"exercise_id":`+itoa(exerciseID)+`,"weight":100,"reps":5,"rpe":8.5}`)
	if setResp.Code != http.StatusCreated {
		t.Fatalf("create set status %d: %s", setResp.Code, setResp.Body.String())
	}

	statsResp := httptest.NewRecorder()
	statsReq := httptest.NewRequest(http.MethodGet, "/users/"+itoa(userID)+"/stats", nil)
	statsReq.Header.Set("Authorization", "Bearer test-api-token")
	server.ServeHTTP(statsResp, statsReq)
	if statsResp.Code != http.StatusOK {
		t.Fatalf("get stats status %d: %s", statsResp.Code, statsResp.Body.String())
	}
	if !strings.Contains(statsResp.Body.String(), `"total_volume":500`) {
		t.Fatalf("expected total volume 500, got %s", statsResp.Body.String())
	}
}

func TestTelegramRoutesRequireInternalToken(t *testing.T) {
	server := newTestServer()

	resp := doJSON(server, http.MethodPost, "/telegram/users", `{"telegram_id":777,"username":"tg"}`)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status %d, got %d: %s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"code":"invalid_internal_token"`) {
		t.Fatalf("expected invalid_internal_token code, got %s", resp.Body.String())
	}
}

func TestTelegramWorkoutFlowDoesNotAcceptUserID(t *testing.T) {
	server := newTestServer()

	userResp := doTelegramJSON(server, http.MethodPost, "/telegram/users", `{"telegram_id":777,"username":"tg"}`)
	if userResp.Code != http.StatusOK {
		t.Fatalf("get or create telegram user status %d: %s", userResp.Code, userResp.Body.String())
	}
	userID := decodeID(t, userResp.Body.Bytes())

	workoutResp := doTelegramJSON(server, http.MethodPost, "/telegram/users/777/workouts", `{"username":"tg","description":"Telegram workout"}`)
	if workoutResp.Code != http.StatusCreated {
		t.Fatalf("create telegram workout status %d: %s", workoutResp.Code, workoutResp.Body.String())
	}

	var workout struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.Unmarshal(workoutResp.Body.Bytes(), &workout); err != nil {
		t.Fatalf("decode workout: %v", err)
	}
	if workout.UserID != userID {
		t.Fatalf("expected workout user_id %d, got %d", userID, workout.UserID)
	}
}

func TestTelegramSetRejectsForeignWorkout(t *testing.T) {
	server := newTestServer()

	ownerResp := doTelegramJSON(server, http.MethodPost, "/telegram/users", `{"telegram_id":1001,"username":"owner"}`)
	if ownerResp.Code != http.StatusOK {
		t.Fatalf("create owner status %d: %s", ownerResp.Code, ownerResp.Body.String())
	}

	foreignResp := doTelegramJSON(server, http.MethodPost, "/telegram/users", `{"telegram_id":1002,"username":"foreign"}`)
	if foreignResp.Code != http.StatusOK {
		t.Fatalf("create foreign status %d: %s", foreignResp.Code, foreignResp.Body.String())
	}

	exerciseResp := doJSON(server, http.MethodPost, "/exercises", `{"name":"Telegram squat"}`)
	if exerciseResp.Code != http.StatusCreated {
		t.Fatalf("create exercise status %d: %s", exerciseResp.Code, exerciseResp.Body.String())
	}
	exerciseID := decodeID(t, exerciseResp.Body.Bytes())

	workoutResp := doTelegramJSON(server, http.MethodPost, "/telegram/users/1001/workouts", `{"username":"owner","description":"Owner workout"}`)
	if workoutResp.Code != http.StatusCreated {
		t.Fatalf("create owner workout status %d: %s", workoutResp.Code, workoutResp.Body.String())
	}
	workoutID := decodeID(t, workoutResp.Body.Bytes())

	setResp := doTelegramJSON(server, http.MethodPost, "/telegram/users/1002/sets", `{"username":"foreign","workout_id":`+itoa(workoutID)+`,"exercise_id":`+itoa(exerciseID)+`,"weight":100,"reps":5,"rpe":8}`)
	if setResp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status %d, got %d: %s", http.StatusForbidden, setResp.Code, setResp.Body.String())
	}
	if !strings.Contains(setResp.Body.String(), `"code":"forbidden"`) {
		t.Fatalf("expected forbidden code, got %s", setResp.Body.String())
	}
}

func newTestServer() http.Handler {
	userRepo := memory.NewUserRepository()
	workoutRepo := memory.NewWorkoutRepository()
	setRepo := memory.NewSetRepository()
	exerciseRepo := memory.NewExerciseRepository()

	h := handler.NewHandler(
		service.NewUserService(userRepo),
		service.NewWorkoutService(workoutRepo, userRepo),
		service.NewSetService(setRepo, workoutRepo, exerciseRepo),
		service.NewStatsService(setRepo, workoutRepo, userRepo),
		service.NewExerciseService(exerciseRepo),
	)

	return router.NewRouter(h, "memory", "test-token", "test-api-token")
}

func doJSON(server http.Handler, method, path, body string) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-api-token")
	server.ServeHTTP(resp, req)
	return resp
}

func doTelegramJSON(server http.Handler, method, path, body string) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", "test-token")
	server.ServeHTTP(resp, req)
	return resp
}

func decodeID(t *testing.T, body []byte) int64 {
	t.Helper()

	var data struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("decode response id: %v", err)
	}
	return data.ID
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
