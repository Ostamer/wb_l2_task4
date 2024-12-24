package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Определение структуры для хранения параметров
type Params struct {
	Column         int
	Digitals       bool
	Reverse        bool
	Unique         bool
	Month          bool
	IgnoreSpace    bool
	CheckSorted    bool
	SuficsDigitals bool
	Delimiter      string
	InputFile      string
	OutputFile     string
}

// Основная функция
func main() {
	params := setParams()
	lines := readFile(params.InputFile)

	if params.CheckSorted {
		if checkSorted(lines, params) {
			fmt.Println("Данные уже отсортированы.")
		} else {
			fmt.Println("Данные не отсортированы.")
		}
		return
	}

	lines = sortLines(lines, params)
	writeOutput(lines, params.OutputFile)
}

// Функция для обработки введеной строки
func setParams() Params {
	column := flag.Int("k", 1, "Указать колонку для сортировки (начиная с 1)")
	digitals := flag.Bool("n", false, "Сортировать по числовому значению")
	reverse := flag.Bool("r", false, "Сортировать в обратном порядке")
	unique := flag.Bool("u", false, "Убрать повторяющиеся строки")
	month := flag.Bool("M", false, "Сортировать по названию месяца")
	ignoreSpace := flag.Bool("b", false, "Игнорировать хвостовые пробелы")
	checkSorted := flag.Bool("c", false, "Проверять, отсортированы ли данные")
	suficsDigitals := flag.Bool("h", false, "Сортировать по числовому значению с учетом суффиксов")
	delimiter := flag.String("d", " ", "Указать разделитель колонок")
	inputFile := flag.String("i", "", "Указать имя входного файла")
	outputFile := flag.String("o", "", "Указать имя выходного файла")

	flag.Parse()

	return Params{
		Column:         *column - 1,
		Digitals:       *digitals,
		Reverse:        *reverse,
		Unique:         *unique,
		Month:          *month,
		IgnoreSpace:    *ignoreSpace,
		CheckSorted:    *checkSorted,
		SuficsDigitals: *suficsDigitals,
		Delimiter:      *delimiter,
		InputFile:      *inputFile,
		OutputFile:     *outputFile,
	}
}

// Фукнция для чтения файла
func readFile(inputFile string) []string {
	var scanner *bufio.Scanner
	if inputFile == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ошибка открытия файла:", err)
			os.Exit(1)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения ввода:", err)
		os.Exit(1)
	}
	return lines
}

// Запись данных в файл иначе при отсутствии указанного файла вывод в консоль
func writeOutput(lines []string, outputFile string) {
	if outputFile == "" {
		for _, line := range lines {
			fmt.Println(line)
		}
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Ошибка создания файла:", err)
			os.Exit(1)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		for _, line := range lines {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				fmt.Fprintln(os.Stderr, "Ошибка записи в файл:", err)
				os.Exit(1)
			}
		}
		writer.Flush()
	}
}

// Сортировка строк
func sortLines(lines []string, params Params) []string {
	if params.IgnoreSpace {
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " ")
		}
	}

	if params.Unique {
		lines = uniqueLines(lines)
	}

	sort.SliceStable(lines, func(i, j int) bool {
		iKey := extractKey(lines[i], params)
		jKey := extractKey(lines[j], params)

		var less bool
		if params.Month {
			less = compareMonths(iKey, jKey)
		} else if params.Digitals {
			less = compareDigitals(iKey, jKey)
		} else if params.SuficsDigitals {
			less = compareSuficsDigitals(iKey, jKey)
		} else {
			less = iKey < jKey
		}

		if params.Reverse {
			return !less
		}
		return less
	})

	return lines
}

func extractKey(line string, params Params) string {
	columns := strings.Split(line, params.Delimiter)
	if params.Column >= 0 && params.Column < len(columns) {
		return columns[params.Column]
	}
	return line
}

// Функция для сравнения чисел
func compareDigitals(a, b string) bool {
	aNum, errA := strconv.ParseFloat(a, 64)
	bNum, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return a < b
	}
	return aNum < bNum
}

// Функция для сравнения месяцев
func compareMonths(a, b string) bool {
	monthOrder := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
		"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	return monthOrder[a] < monthOrder[b]
}

// Функция для сравнения двух чисел с суфиксами
func compareSuficsDigitals(a, b string) bool {
	aNum := parseSuficsDigitals(a)
	bNum := parseSuficsDigitals(b)
	return aNum < bNum
}

// Функция для преобразования числа с суфиксом в число
func parseSuficsDigitals(s string) float64 {
	units := map[byte]float64{
		'K': 1e3, 'M': 1e6, 'G': 1e9, 'T': 1e12,
	}
	length := len(s)
	if length == 0 {
		return 0
	}

	if multiplier, ok := units[s[length-1]]; ok {
		value, err := strconv.ParseFloat(s[:length-1], 64)
		if err != nil {
			return 0
		}
		return value * multiplier
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return value
}

// Функиця для определения уникальных строк
func uniqueLines(lines []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

// Функция для определния отсортирован ли файл
func checkSorted(lines []string, params Params) bool {
	for i := 1; i < len(lines); i++ {
		if params.Reverse {
			if extractKey(lines[i-1], params) < extractKey(lines[i], params) {
				return false
			}
		} else {
			if extractKey(lines[i-1], params) > extractKey(lines[i], params) {
				return false
			}
		}
	}
	return true
}
