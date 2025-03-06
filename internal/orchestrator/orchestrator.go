package orchestrator

import (
	models "Calc_2GO/Models"
	"Calc_2GO/Pkg/calculator"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Orchestrator struct {
	mu          sync.Mutex
	expressions map[int]*Expression
	tasks       []models.Task
	results     map[int]float64
}

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		expressions: make(map[int]*Expression),
		tasks:       []models.Task{},
		results:     make(map[int]float64),
	}
}

func (o *Orchestrator) AddExpression(expr string) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := len(o.expressions) + 1
	expression := &Expression{ID: id, Status: "pending"}
	o.expressions[id] = expression

	tasks, err := calculator.CalcToTasks(id, expr)
	if err != nil {
		log.Printf("❌ Ошибка при разборе выражения: %v", err)
		return 0, fmt.Errorf("ошибка при разборе выражения: %v", err)
	}

	o.tasks = append(o.tasks, tasks...)
	log.Printf("✅ Добавлено выражение: %s", expr)
	return id, nil
}

func (o *Orchestrator) GetExpression(id int) (*Expression, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	expr, exists := o.expressions[id]
	return expr, exists
}

func (o *Orchestrator) GetAllExpressions() []*Expression {
	result := make([]*Expression, 0, len(o.expressions))

	for _, expr := range o.expressions {
		result = append(result, expr)
	}

	return result
}

func (o *Orchestrator) GetNextTask() (*models.Task, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.tasks) == 0 {
		log.Println("❌ Нет задач, готовых к выполнению")
		return nil, false
	}

	task := o.tasks[0]
	o.tasks = o.tasks[1:]
	log.Printf("✅ Задача id %d передана агенту. Выражение: %v", task.ID, task)
	return &task, true
}

func (o *Orchestrator) HandleGetExpressions(w http.ResponseWriter, r *http.Request) {
	expressions := o.GetAllExpressions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expressions)
}

func (o *Orchestrator) HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "❌ Неверный формат ID", http.StatusBadRequest)
		return
	}

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

	idStr := strconv.Itoa(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": idStr})
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

	o.expressions[task.ID].Status = "in_progress"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (o *Orchestrator) HandleTaskResult(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     int     `json:"id"`
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
	log.Printf("✅ Результат задачи %d записан: %f", request.ID, request.Result)

	isTaskExist := false
	for _, task := range o.tasks {
		if task.ID == request.ID {
			isTaskExist = true
			break
		}
	}

	if !isTaskExist {
		o.expressions[request.ID].Result = o.results[request.ID]
		o.expressions[request.ID].Status = "done"
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
