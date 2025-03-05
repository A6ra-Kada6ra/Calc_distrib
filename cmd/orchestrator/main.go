package main

import (
	"Calc_2GO/Internal/orchestrator"
	"log"
)

func main() {
	// Создаем новый оркестратор
	o := orchestrator.NewOrchestrator()

	// Запускаем сервер оркестратора
	log.Println("🛠️ Запуск оркестратора...")
	o.StartServer()
}
