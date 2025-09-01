package files

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ctxengine/config"

	"github.com/fsnotify/fsnotify"
)

// WatchAndReindex следит за изменениями в директории проекта, игнорируя .git
func WatchAndReindex(cfg *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Ошибка создания наблюдателя: %v", err)
	}
	defer watcher.Close()

	// Рекурсивно добавляем наблюдение за директорией, игнорируя .git
	err = AddWatchersRecursively(watcher, cfg.Context.ProjectDir)
	if err != nil {
		log.Fatalf("Ошибка добавления наблюдения: %v", err)
	}

	log.Printf("Наблюдение запущено для директории: %s", cfg.Context.ProjectDir)

	// Дебаунс: если в течение N секунд не появилось новых событий, запускаем переиндексацию.
	debounceDelay := time.Duration(cfg.Context.TimeReIndex) * time.Second
	var timer *time.Timer

	resetTimer := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceDelay, func() {
			log.Printf("Нет изменений в течение %d секунд. Запуск переиндексации...", cfg.Context.TimeReIndex)
			ReindexWithWorkerPool(cfg)
		})
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Printf("Обнаружено событие: %s", event)
			resetTimer()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Ошибка наблюдения: %v", err)
		}
	}
}

// AddWatchersRecursively добавляет watcher для всех поддиректорий от корня,
// пропуская директории, имя которых содержит ".git".
func AddWatchersRecursively(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Если это директория и она содержит ".git" в имени, пропускаем её.
		if info.IsDir() {
			// Если имя директории равно ".git", пропускаем её и не идем вглубь.
			if strings.EqualFold(info.Name(), ".git") {
				return filepath.SkipDir
			}
			log.Printf("Добавляем наблюдение за директорией: %s", path)
			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}
