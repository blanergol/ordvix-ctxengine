package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"ctxengine/config"
	"ctxengine/internal/files"
)

func main() {
	// Инициализация конфигурации
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации: %v", err)
	}

	// Первоначальная индексация через pipeline
	files.ReindexWithWorkerPool(cfg)

	// Запускаем горутину, которая следит за изменениями файлов
	go files.WatchAndReindex(cfg)

	// Ожидаем сигнал завершения (например, Ctrl+C)
	WaitForSignal()
}

// WaitForSignal блокирует основной поток до получения сигнала завершения.
func WaitForSignal() {
	// Ожидаем сигнала завершения (SIGINT, SIGTERM) для корректного завершения работы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Приложение запущено. Ожидаем сигнала завершения...")
	sig := <-sigs
	log.Printf("Получен сигнал: %s. Завершаем работу.", sig)
}
