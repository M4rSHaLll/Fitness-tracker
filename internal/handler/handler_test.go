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
	server.ServeHTTP(statsResp, statsReq)
	if statsResp.Code != http.StatusOK {
		t.Fatalf("get stats status %d: %s", statsResp.Code, statsResp.Body.String())
	}
	if !strings.Contains(statsResp.Body.String(), `"total_volume":500`) {
		t.Fatalf("expected total volume 500, got %s", statsResp.Body.String())
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

	return router.NewRouter(h, "memory")
}

func doJSON(server http.Handler, method, path, body string) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
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
