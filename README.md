# 🇳🇱 Dutch Learning Telegram Bot

A sophisticated Telegram bot for learning Dutch vocabulary using spaced repetition and contextual grammar tips. Built with Go, implementing the FSRS (Free Spaced Repetition Scheduler) algorithm for optimal learning retention.

## ✨ Features

### 🧠 Smart Learning System
- **Spaced Repetition**: FSRS v4 algorithm with 90% target retention
- **Contextual Grammar Tips**: Smart tips that appear only when relevant to the current word
- **Adaptive Difficulty**: Questions adapt based on your performance
- **Multiple Choice Format**: User-friendly multiple choice questions
- **Progress Tracking**: Detailed statistics and learning analytics

### 🎯 Contextual Grammar Intelligence
- **Category-Based Tips**: Tips designed for specific word categories (home, body, verbs, etc.)
- **Pattern Recognition**: Tips trigger based on word patterns (compound words, verb endings, etc.)
- **Smart Probability**: 20% chance of showing relevant tips during learning
- **Real Examples**: Each tip includes Dutch and English examples

### 📊 Advanced Features
- **User Preferences**: Toggle grammar tips and smart reminders
- **Smart Reminders**: Intelligent notifications during active hours (8 AM - 10 PM)
- **Anti-Repetition**: Prevents showing the same words too frequently
- **Bidirectional Learning**: Both Dutch→English and English→Dutch questions
- **Performance Analytics**: Track your learning progress and retention rates

### 🏠 Rich Vocabulary Database
- **500+ Words** across multiple categories:
  - Home & Household items
  - Body parts
  - Family & People
  - Food & Drinks
  - Nature & Animals
  - Colors & Numbers
  - And more...

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- SQLite (automatically handled)

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/DSempai/langbot.git
cd langbot
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Set up environment:**
```bash
cp .env.example .env
# Edit .env and add your Telegram Bot Token
```

4. **Build and run:**
```bash
go build -o langbot cmd/bot/main.go
./langbot
```

### Environment Configuration

Create a `.env` file with:
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_PATH=dutch_learning.db
LOG_LEVEL=info
```

## 🎮 How to Use

### Getting Started
1. Start a chat with your bot on Telegram
2. Send `/start` to begin
3. Choose "📚 Start Learning" from the menu
4. Answer questions and learn Dutch!

### Settings & Customization
- **⚙️ Settings**: Access via main menu
- **🎯 Grammar Tips**: Toggle contextual grammar guidance
- **🔔 Smart Reminders**: Enable/disable learning reminders
- **📊 Statistics**: View your learning progress

### Learning Flow
1. **Question Presentation**: You'll see a word to translate
2. **Multiple Choice**: Select from 4 options
3. **Immediate Feedback**: Know if you're right or wrong
4. **Grammar Tips**: Occasionally get relevant grammar insights
5. **Spaced Repetition**: Words reappear based on your performance

## 🏗️ Technical Architecture

### Domain-Driven Design (DDD)
The project follows DDD principles with clear separation of concerns:

```
internal/
├── domain/                 # Core business logic
│   ├── entities/          # Domain entities (User, Word, etc.)
│   ├── repositories/      # Repository interfaces
│   └── services/          # Domain services
├── application/           # Application layer
│   └── usecases/         # Use cases and business workflows
├── infrastructure/       # External concerns
│   └── persistence/      # Database implementations
└── interfaces/           # Interface adapters
    └── telegram/         # Telegram bot handlers
```

### Key Components

#### 🗄️ Database Schema
- **Users**: User profiles and preferences
- **Words**: Vocabulary with categories and metadata
- **User Words**: Learning progress and FSRS scheduling
- **Grammar Tips**: Contextual grammar guidance system
- **User Preferences**: Flexible settings system

#### 🤖 FSRS Integration
- **Optimal Intervals**: Scientifically-backed spacing intervals
- **Difficulty Tracking**: Cards track difficulty and stability
- **Performance Adaptation**: Intervals adjust based on user performance

#### 🎯 Grammar Tips System
- **Contextual Matching**: Tips match current vocabulary
- **Pattern Recognition**: Smart triggering based on word patterns
- **Category Awareness**: Tips designed for specific word categories

## 📈 Grammar Tips Examples

The bot includes intelligent grammar tips that appear contextually:

- **Learning "slaapkamer"** → Tip about compound words with -kamer
- **Learning "koelkast"** → Tip about compound appliances  
- **Learning "werk"** → Tip about Dutch verb past tense formation
- **Learning body parts** → Tip about articles with body parts
- **Learning locations** → Tip about prepositions (in/op/aan)

## 🔧 Development

### Project Structure
```
langbot/
├── cmd/bot/main.go           # Application entry point
├── internal/                 # Private application code
├── vocabulary.json           # Vocabulary database
├── grammar_tips.json         # Grammar tips database
├── go.mod                   # Go module definition
└── README.md                # This file
```

### Adding New Features

#### Adding Vocabulary
Edit `vocabulary.json`:
```json
{
  "word": "new_word",
  "translation": "nieuwe_woord", 
  "category": "category_name"
}
```

#### Adding Grammar Tips
Edit `grammar_tips.json`:
```json
{
  "title": "Tip Title",
  "explanation": "Grammar explanation...",
  "dutch_example": "Dutch example",
  "english_example": "English example", 
  "category": "grammar_category",
  "applicable_categories": ["word_category"],
  "specific_words": ["specific", "words"]
}
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -ldflags="-s -w" -o langbot cmd/bot/main.go
```

## 📊 Statistics & Analytics

The bot tracks comprehensive learning statistics:
- **Total words learned**
- **Daily/Weekly progress**
- **Retention rates**
- **Learning streaks**
- **Category-specific progress**
- **Grammar tips engagement**

## 🔐 Privacy & Security

- **Local Database**: All user data stored locally in SQLite
- **No External APIs**: Except Telegram Bot API
- **Minimal Data Collection**: Only learning progress and preferences
- **Open Source**: Full transparency in code and data handling

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go conventions and best practices
- Maintain DDD architecture principles
- Add tests for new features
- Update documentation as needed

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **FSRS Algorithm**: Based on the Free Spaced Repetition Scheduler
- **Dutch Language Resources**: Various online Dutch learning materials
- **Telegram Bot API**: For the excellent bot platform
- **Go Community**: For the robust standard library and ecosystem

## 📬 Support

- **Issues**: Report bugs via [GitHub Issues](https://github.com/DSempai/langbot/issues)
- **Features**: Request features via [GitHub Discussions](https://github.com/DSempai/langbot/discussions)
- **Contact**: Reach out via GitHub

---

**Happy Learning!** 🇳🇱📚

*Made with ❤️ for Dutch language learners* 