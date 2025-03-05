package main

import (
	"Calc_2GO/internal/agent"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// URL оркестратора
	orchestratorURL := "http://localhost:8080"

	// Количество горутин (вычислительных мощностей)
	computingPower := 2

	// Получаем время выполнения операций из переменных среды
	timeAddition := getEnvDuration("TIME_ADDITION_MS")
	timeSubtraction := getEnvDuration("TIME_SUBTRACTION_MS")
	timeMultiplication := getEnvDuration("TIME_MULTIPLICATION_MS")
	timeDivision := getEnvDuration("TIME_DIVISION_MS")

	// Создаем агента
	agent := agent.NewAgent(orchestratorURL, computingPower, timeAddition, timeSubtraction, timeMultiplication, timeDivision)

	// Запуск агента
	log.Println("🚀 Запуск агента...")
	agent.Start()

	// Бесконечное ожидание (чтобы программа не завершилась)
	select {}
}

// Функция для получения времени выполнения операций из переменных среды
func getEnvDuration(key string) time.Duration {
	var defaultValue time.Duration = 1000
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

// FIXME: set: os.Setenv("TIME_ADDITION_MS", "5000")
