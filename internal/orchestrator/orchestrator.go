package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Orchestrator struct {
	mu          sync.Mutex
	expressions map[string]*Expression
	tasks       []Task
	results     map[string]float64
}

type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type Task struct {
	ID            string        `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		expressions: make(map[string]*Expression),
		tasks:       []Task{},
		results:     make(map[string]float64),
	}
}

func (o *Orchestrator) AddExpression(expr string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := "1"
	return id, nil
}

func (o *Orchestrator) GetExpression(id string) (*Expression, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	expr, exists := o.expressions[id]
	return expr, exists
}

func (o *Orchestrator) GetAllExpressions() []*Expression {
	o.mu.Lock()
	defer o.mu.Unlock()

	expressions := []*Expression{
		{ID: "1", Status: "pending", Result: 1.0},
		{ID: "2", Status: "completed", Result: 42.0},
	}

	return expressions
}

func (o *Orchestrator) GetNextTask() (*Task, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Println("❌ Нет задач, готовых к выполнению")
	return nil, false
}
func (o *Orchestrator) HandleGetExpressions(w http.ResponseWriter, r *http.Request) {
	expressions := o.GetAllExpressions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expressions)
}

func (o *Orchestrator) HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	expr, exists := o.GetExpression(id)
	if !exists {
		http.Error(w, "❌ Выражение не найдено", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expr)
}

func (o *Orchestrator) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("❌ Ошибка при чтении данных: %v", err), http.StatusBadRequest)
		return
	}

	id, err := o.AddExpression(request.Expression)
	if err != nil {
		http.Error(w, fmt.Sprintf("❌ Ошибка при добавлении выражения: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (o *Orchestrator) HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		o.HandleTaskResult(w, r)
		return
	default:

	}

	task, exists := o.GetNextTask()
	if !exists {
		http.Error(w, "❌ Нет доступных задач", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (o *Orchestrator) HandleTaskResult(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Ошибка при чтении данных: %v", err)
		http.Error(w, fmt.Sprintf("❌ Ошибка при чтении данных: %v", err), http.StatusBadRequest)
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.results[request.ID] = request.Result
	log.Printf("✅ Результат задачи %s записан: %f", request.ID, request.Result)

	for _, expr := range o.expressions {
		if expr.Status == "pending" || expr.Status == "in_progress" {
			allTasksCompleted := true
			for _, task := range o.tasks {
				if strings.HasPrefix(task.ID, expr.ID) {
					allTasksCompleted = false
					break
				}
			}

			if allTasksCompleted {
				expr.Status = "completed"
				expr.Result = o.results[expr.ID+"-final"]
				log.Printf("✅ Выражение %s завершено, результат: %f", expr.ID, expr.Result)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
func (o *Orchestrator) StartServer() {
	fmt.Println("🚀 Оркестратор запущен на порту 8080")

	http.HandleFunc("/api/v1/calculate", o.HandleCalculate)
	http.HandleFunc("/api/v1/expressions", o.HandleGetExpressions)
	http.HandleFunc("/api/v1/expressions/", o.HandleGetExpressionByID)
	http.HandleFunc("/internal/task", o.HandleTask)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("❌ Ошибка запуска сервера:", err)
	}
}
