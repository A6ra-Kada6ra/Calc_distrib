package main

import (
	"Calc_2GO/Internal/agent"
	"log"
)

func main() {
	// URL оркестратора
	orchestratorURL := "http://localhost:8080"

	// Количество горутин (вычислительных мощностей)
	computingPower := 2

	// Создаем агента
	agent := agent.NewAgent(orchestratorURL, computingPower)

	// Запуск агента
	log.Println("🚀 Запуск агента...")
	agent.Start()

	// Бесконечное ожидание (чтобы программа не завершилась)
	select {}
}
