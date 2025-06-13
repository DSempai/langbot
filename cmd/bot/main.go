package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dutch-learning-bot/internal/application/usecases"
	"dutch-learning-bot/internal/infrastructure/filesystem"
	"dutch-learning-bot/internal/infrastructure/persistence"
	"dutch-learning-bot/internal/infrastructure/telegram"
	"dutch-learning-bot/internal/interfaces/telegram/handlers"
)

func main() {
	// Get bot token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dutch_learning.db"
	}
	db, err := persistence.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := persistence.NewUserRepository(db)
	preferencesRepo := persistence.NewUserPreferencesRepository(db)
	vocabularyRepo := persistence.NewVocabularyRepository(db)
	learningRepo := persistence.NewLearningRepository(db)
	grammarRepo := persistence.NewGrammarRepository(db)

	// Load and populate vocabulary
	vocabularyLoader := filesystem.NewVocabularyLoader()
	vocabulary, err := vocabularyLoader.LoadFromFile("vocabulary.json")
	if err != nil {
		log.Fatalf("Failed to load vocabulary: %v", err)
	}

	err = vocabularyRepo.SaveBatch(context.Background(), vocabulary)
	if err != nil {
		log.Fatalf("Failed to populate vocabulary: %v", err)
	}

	// Load and populate grammar tips
	grammarLoader := filesystem.NewGrammarLoader()
	grammarTips, err := grammarLoader.LoadFromFile("grammar_tips.json")
	if err != nil {
		log.Fatalf("Failed to load grammar tips: %v", err)
	}

	err = grammarRepo.SaveBatch(context.Background(), grammarTips)
	if err != nil {
		log.Fatalf("Failed to populate grammar tips: %v", err)
	}

	// Initialize use cases
	userUseCase := usecases.NewUserUseCase(userRepo, preferencesRepo)
	learningUseCase := usecases.NewLearningUseCase(learningRepo, vocabularyRepo, userRepo, grammarRepo, preferencesRepo)

	// Initialize Telegram bot
	bot, err := telegram.NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Setup bot commands with Telegram
	if err := bot.SetupCommands(); err != nil {
		log.Printf("Warning: Failed to setup bot commands: %v", err)
		log.Printf("The bot will still work, but commands won't show in Telegram's menu")
	}

	// Initialize reminder service
	reminderUseCase := usecases.NewReminderUseCase(bot, userRepo, learningRepo, preferencesRepo, nil)

	// Initialize handler
	handler := handlers.NewBotHandler(bot, userUseCase, learningUseCase, preferencesRepo)

	// Start bot
	log.Printf("Starting Dutch Learning Bot...")

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start reminder service in background
	go reminderUseCase.StartReminderService(ctx)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	if err := handler.Start(ctx); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
