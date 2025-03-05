package calculator_test

import (
	"Calc_distrib/pkg/calculator"
	"fmt"
	"testing"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
		wantErr    bool
		errMsg     string
	}{
		{"Простая операция сложения", "2+2", 4, false, ""},
		{"Простая операция умножения", "3*3", 9, false, ""},
		{"Приоритет умножения над сложением", "2+2*2", 6, false, ""},
		{"Комбинация деления и вычитания", "10/2-1", 4, false, ""},
		{"Смешанная операция", "5-1+3*2", 10, false, ""},
		{"Скобки с приоритетом", "(2+3)*4", 20, false, ""},
		{"Скобки внутри скобок", "((1+2)*3)", 9, false, ""},
		{"Сложное выражение со скобками", "(2+(2*2)+(3+4))*2", 26, false, ""},
		{"Деление на ноль", "10/0", 0, true, "division by zero"},
		{"Нечисловой токен", "3^2", 0, true, "invalid character: 3^2"},
		{"Пустое выражение", "", 0, true, "invalid expression"},
		{"Несбалансированные скобки", "(2+3", 0, true, "mismatched parentheses"},
		{"Неверный символ", "2 + a", 0, true, "invalid character: a"},
		{"Отрицательные числа", "-2+3", 1, false, ""},
		{"Десятичные числа", "3.5+2.5", 6, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.Calc(tt.expression)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("❌ %s: ожидалась ошибка, но получили результат: %v", tt.name, got)
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Fatalf("❌ %s: ожидали сообщение об ошибке '%s', а получили '%s'", tt.name, tt.errMsg, err.Error())
				}
				fmt.Printf("✅ %s: корректно отловлена ошибка '%s'\n", tt.name, err.Error())
			} else {
				if err != nil {
					t.Fatalf("❌ %s: не ожидали ошибку, но получили: %v", tt.name, err)
				}
				if got != tt.want {
					t.Fatalf("❌ %s: ожидали %g, а получили %g", tt.name, tt.want, got)
				}
				fmt.Printf("✅ %s: %s = %g\n", tt.name, tt.expression, got)
			}
		})
	}
}
