package orchestrator_test

import (
	"Calc_distrib/internal/orchestrator"
	"encoding/json"
	"fmt"
	"io" // Добавлен импорт
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOrchestrator(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantStatus string
		wantResult float64
		wantErr    bool
		errMsg     string
	}{
		{"Простое выражение", "2+2", "completed", 4, false, ""},
		{"Приоритет операций", "2+2*2", "completed", 6, false, ""},
		{"Скобки", "(2+3)*4", "completed", 20, false, ""},
		{"Деление на ноль", "10/0", "completed", 0, true, "division by zero"},
		{"Неизвестная операция", "2^3", "completed", 0, true, "unknown operation"},
		{"Пустое выражение", "", "pending", 0, true, "invalid expression"},
		{"Несбалансированные скобки", "(2+3", "pending", 0, true, "mismatched parentheses"},
		{"Неверный символ", "2 + a", "pending", 0, true, "invalid character: a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := orchestrator.NewOrchestrator()

			// Добавляем выражение
			id, err := o.AddExpression(tt.expression)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("❌ %s: ожидалась ошибка, но её нет", tt.name)
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Fatalf("❌ %s: ожидали сообщение об ошибке '%s', а получили '%s'", tt.name, tt.errMsg, err.Error())
				}
				fmt.Printf("✅ %s: корректно отловлена ошибка '%s'\n", tt.name, err.Error())
				return
			}

			if err != nil {
				t.Fatalf("❌ %s: не ожидали ошибку, но получили: %v", tt.name, err)
			}

			// Получаем задачу
			task, exists := o.GetNextTask()
			if !exists {
				t.Fatalf("❌ %s: задача не найдена", tt.name)
			}

			// Выполняем задачу
			result, err := executeTask(task)
			if err != nil {
				t.Fatalf("❌ %s: ошибка при выполнении задачи: %v", tt.name, err)
			}

			// Отправляем результат
			o.HandleTaskResult(httptest.NewRecorder(), &http.Request{
				Body: io.NopCloser(jsonBody(map[string]interface{}{"id": task.ID, "result": result})), // Исправлено
			})

			// Проверяем статус выражения
			expr, exists := o.GetExpression(id)
			if !exists {
				t.Fatalf("❌ %s: выражение не найдено", tt.name)
			}

			if expr.Status != tt.wantStatus {
				t.Fatalf("❌ %s: ожидали статус '%s', а получили '%s'", tt.name, tt.wantStatus, expr.Status)
			}

			if expr.Result != tt.wantResult {
				t.Fatalf("❌ %s: ожидали результат %g, а получили %g", tt.name, tt.wantResult, expr.Result)
			}

			fmt.Printf("✅ %s: выражение '%s' выполнено успешно, результат: %g\n", tt.name, tt.expression, expr.Result)
		})
	}
}

func executeTask(task *orchestrator.Task) (float64, error) {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("unknown operation")
	}
}

func jsonBody(data map[string]interface{}) *strings.Reader {
	body, _ := json.Marshal(data)
	return strings.NewReader(string(body))
}
