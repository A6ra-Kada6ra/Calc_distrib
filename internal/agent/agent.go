package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Task struct {
	ID            string        `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
	DependsOn     []string      `json:"depends_on"`
}

type Agent struct {
	orchestratorURL    string
	computingPower     int
	timeAddition       time.Duration
	timeSubtraction    time.Duration
	timeMultiplication time.Duration
	timeDivision       time.Duration
	logger             *log.Logger
}

func NewAgent(orchestratorURL string, computingPower int, timeAddition, timeSubtraction, timeMultiplication, timeDivision time.Duration) *Agent {
	if computingPower <= 0 {
		computingPower = 1
	}
	logger := log.New(os.Stdout, "[AGENT] ", log.LstdFlags)
	return &Agent{
		orchestratorURL:    orchestratorURL,
		computingPower:     computingPower,
		timeAddition:       timeAddition,
		timeSubtraction:    timeSubtraction,
		timeMultiplication: timeMultiplication,
		timeDivision:       timeDivision,
		logger:             logger,
	}
}

func (a *Agent) Start() {
	for i := 0; i < a.computingPower; i++ {
		go a.worker()
	}
}

func (a *Agent) worker() {
	for {
		task, err := a.getTask()
		if err != nil {
			a.logger.Printf("❌ ошибка при получении задачи: %v\n", err) // Исправлено
			time.Sleep(2 * time.Second)
			continue
		}

		result, err := a.ExecuteTask(task)
		if err != nil {
			a.logger.Printf("❌ ошибка при выполнении задачи %s: %v\n", task.ID, err) // Исправлено
			continue
		}

		if err := a.submitTaskResult(task.ID, result); err != nil {
			a.logger.Printf("❌ ошибка при отправке результата задачи %s: %v\n", task.ID, err) // Исправлено
		} else {
			a.logger.Printf("✅ результат задачи %s успешно отправлен: %f\n", task.ID, result)
		}
	}
}

func (a *Agent) getTask() (*Task, error) {
	resp, err := http.Get(a.orchestratorURL + "/internal/task")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе задачи: %w", err) // Исправлено
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("задачи недоступны, код ответа: %d", resp.StatusCode) // Исправлено
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании задачи: %w", err) // Исправлено
	}
	a.logger.Printf("задача %s получена", task.ID)
	return &task, nil
}

func (a *Agent) ExecuteTask(task *Task) (float64, error) {
	a.logger.Printf("выполнение задачи %s: %f %s %f", task.ID, task.Arg1, task.Operation, task.Arg2)

	var operationTime time.Duration
	switch task.Operation {
	case "+":
		operationTime = a.timeAddition
	case "-":
		operationTime = a.timeSubtraction
	case "*":
		operationTime = a.timeMultiplication
	case "/":
		operationTime = a.timeDivision

	default:
		return 0, fmt.Errorf("неизвестная операция: %s", task.Operation)
	}

	time.Sleep(operationTime)

	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		result = task.Arg1 / task.Arg2
	default:
		return 0, fmt.Errorf("неизвестная операция: %s", task.Operation)
	}

	a.logger.Printf("✅ задача %s выполнена, результат: %f", task.ID, result)
	return result, nil
}

func (a *Agent) submitTaskResult(taskID string, result float64) error {
	req := struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}

	reqBody, _ := json.Marshal(req)
	a.logger.Printf("отправка результата задачи %s: %f", taskID, result)

	resp, err := http.Post(a.orchestratorURL+"/internal/task", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("ошибка при отправке результата: %w", err) // Исправлено
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("не удалось отправить результат, код ответа: %d", resp.StatusCode) // Исправлено
	}
	return nil
}
