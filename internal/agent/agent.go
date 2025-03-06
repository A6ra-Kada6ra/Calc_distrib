package agent

import (
	models "Calc_2GO/Models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	orchestratorURL    string
	computingPower     int
	timeAddition       time.Duration
	timeSubtraction    time.Duration
	timeMultiplication time.Duration
	timeDivision       time.Duration
	logger             *log.Logger
	taskQueue          chan *models.Task
	wg                 sync.WaitGroup
}

func NewAgent(orchestratorURL string, computingPower int) *Agent {
	os.Setenv("TIME_ADDITION_MS", "10_000")
	os.Setenv("TIME_SUBTRACTION_MS", "10_000")
	os.Setenv("TIME_MULTIPLICATION_MS", "10_000")
	os.Setenv("TIME_DIVISION_MS", "10_000")

	if computingPower <= 0 {
		computingPower = 1
	}

	logger := log.New(os.Stdout, "[AGENT] ", log.LstdFlags)

	timeAddition := getEnvDuration("TIME_ADDITION_MS")
	timeSubtraction := getEnvDuration("TIME_SUBTRACTION_MS")
	timeMultiplication := getEnvDuration("TIME_MULTIPLICATION_MS")
	timeDivision := getEnvDuration("TIME_DIVISION_MS")

	return &Agent{
		orchestratorURL:    orchestratorURL,
		computingPower:     computingPower,
		timeAddition:       timeAddition,
		timeSubtraction:    timeSubtraction,
		timeMultiplication: timeMultiplication,
		timeDivision:       timeDivision,
		logger:             logger,
		taskQueue:          make(chan *models.Task, computingPower),
	}
}

func (a *Agent) Start() {
	for i := 0; i < a.computingPower; i++ {
		go a.worker(i)
	}

	go a.taskDispatcher()
}

func (a *Agent) taskDispatcher() {
	for {
		task, err := a.getTask()
		if err != nil {
			a.logger.Printf("❌ ошибка при получении задачи: %v\n", err) // Исправлено
			time.Sleep(2 * time.Second)
			continue
		}

		a.wg.Add(1)
		a.taskQueue <- task
		a.wg.Wait()
	}
}

func (a *Agent) worker(id int) {
	for task := range a.taskQueue {
		a.logger.Printf("✅ Агент №%d результат задачи %d успешно отправлен:", id, task.ID)

		result, err := a.ExecuteTask(task)
		if err != nil {
			a.logger.Printf("❌ ошибка при выполнении задачи %d: %v\n", task.ID, err) // Исправлено
			a.taskQueue <- task
			continue
		}

		if err := a.submitTaskResult(task.ID, result); err != nil {
			a.logger.Printf("❌ ошибка при отправке результата задачи %d: %v\n", task.ID, err) // Исправлено
			a.taskQueue <- task
		} else {
			a.logger.Printf("✅ Агент №%d результат задачи %d успешно отправлен: %f\n", id, task.ID, result)
		}

		a.wg.Done()
	}
}

func (a *Agent) getTask() (*models.Task, error) {
	resp, err := http.Get(a.orchestratorURL + "/internal/task")
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе задачи: %w", err) // Исправлено
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("задачи недоступны, код ответа: %d", resp.StatusCode) // Исправлено
	}

	var task models.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании задачи: %w", err) // Исправлено
	}
	a.logger.Printf("задача %d получена", task.ID)
	return &task, nil
}

func (a *Agent) ExecuteTask(task *models.Task) (float64, error) {
	a.logger.Printf("выполнение задачи %d: %f %s %f", task.ID, task.Arg1, task.Operation, task.Arg2)

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

	a.logger.Printf("✅ задача %d выполнена, результат: %f", task.ID, result)
	return result, nil
}

func (a *Agent) submitTaskResult(taskID int, result float64) error {
	req := struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}

	reqBody, _ := json.Marshal(req)
	a.logger.Printf("отправка результата задачи %d: %f", taskID, result)

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

func getEnvDuration(key string) time.Duration {
	var defaultValue time.Duration = 2_000
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	ms, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return time.Duration(ms)
}
