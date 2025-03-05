package calculator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrMismatchedParens  = errors.New("mismatched parentheses")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidCharacter  = errors.New("invalid character")
)

func Calc(expression string) (float64, error) {
	if expression == "" {
		return 0, ErrInvalidExpression
	}

	tokens, err := tokenize(expression)
	if err != nil {
		return 0, err
	}

	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return 0, err
	}

	result, err := evaluatePostfix(postfix)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func tokenize(expr string) ([]string, error) {
	var tokens []string
	var currentToken strings.Builder

	for i, char := range expr {
		if char == ' ' {
			continue
		} else if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			if char == '-' && (i == 0 || expr[i-1] == '(') {
				currentToken.WriteRune(char)
			} else {
				tokens = append(tokens, string(char))
			}
		} else {
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens, nil
}

func infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, ErrMismatchedParens
			}
			operators = operators[:len(operators)-1]
		} else if isOperator(token) {
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(token) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		} else {
			return nil, fmt.Errorf("%w: %s", ErrInvalidCharacter, token)
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return nil, ErrMismatchedParens
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}
	return output, nil
}

func evaluatePostfix(postfix []string) (float64, error) {
	var stack []float64

	for _, token := range postfix {
		if isNumber(token) {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, ErrInvalidToken
			}
			stack = append(stack, num)
		} else if isOperator(token) {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, ErrDivisionByZero
				}
				stack = append(stack, a/b)
			}
		} else {
			return 0, fmt.Errorf("%w: %s", ErrInvalidToken, token)
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}
	return stack[0], nil
}

func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}
