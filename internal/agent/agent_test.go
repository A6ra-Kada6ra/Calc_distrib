package agent_test

import (
	models "Calc_2GO/Models"
	"Calc_2GO/internal/agent"
	"encoding/json" // Добавлен импорт
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgent(t *testing.T) {
	tests := []struct {
		name       string
		task       models.Task
		wantResult float64
		wantErr    bool
		errMsg     string
	}{
		{"Сложение", models.Task{ID: 1, Arg1: 2, Arg2: 2, Operation: "+", OperationTime: 1 * time.Second}, 4, false, ""},
		{"Вычитание", models.Task{ID: 2, Arg1: 5, Arg2: 3, Operation: "-", OperationTime: 1 * time.Second}, 2, false, ""},
		{"Умножение", models.Task{ID: 3, Arg1: 3, Arg2: 3, Operation: "*", OperationTime: 1 * time.Second}, 9, false, ""},
		{"Деление", models.Task{ID: 4, Arg1: 10, Arg2: 2, Operation: "/", OperationTime: 1 * time.Second}, 5, false, ""},
		{"Деление на ноль", models.Task{ID: 5, Arg1: 10, Arg2: 0, Operation: "/", OperationTime: 1 * time.Second}, 0, true, "деление на ноль"},
		{"Неизвестная операция", models.Task{ID: 6, Arg1: 2, Arg2: 2, Operation: "^", OperationTime: 1 * time.Second}, 0, true, "неизвестная операция: ^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Мок сервера оркестратора
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/internal/task" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tt.task)
				} else if r.URL.Path == "/internal/task/result" {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer ts.Close()

			// Создаем агента
			ag := agent.NewAgent(ts.URL, 2)

			// Выполняем задачу
			result, err := ag.ExecuteTask(&tt.task)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("❌ %s: ожидалась ошибка, но её нет", tt.name)
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Fatalf("❌ %s: ожидали сообщение об ошибке '%s', а получили '%s'", tt.name, tt.errMsg, err.Error())
				}
				fmt.Printf("✅ %s: корректно отловлена ошибка '%s'\n", tt.name, err.Error())
			} else {
				if err != nil {
					t.Fatalf("❌ %s: не ожидали ошибку, но получили: %v", tt.name, err)
				}
				if result != tt.wantResult {
					t.Fatalf("❌ %s: ожидали результат %g, а получили %g", tt.name, tt.wantResult, result)
				}
				fmt.Printf("✅ %s: задача выполнена успешно, результат: %g\n", tt.name, result)
			}
		})
	}
}
