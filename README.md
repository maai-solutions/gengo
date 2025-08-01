# GenGo

GenGo is a comprehensive tool for testing generative AI systems using Go.

## Features

### 🎯 Interactive CLI with Bubble Tea
- Beautiful terminal UI with real-time input and cursor
- Type `/exit` or `Ctrl+C` to quit
- Command history display

### 📄 PDF Text Extraction
- Extract text from PDF files using multiple input/output formats:
  - File path → String/Bytes/Writer
  - Byte array → String/Bytes/Writer  
  - Reader → String/Bytes/Writer
- Support for specific page extraction
- Text cleaning and normalization
- PDF information retrieval (page count, etc.)

### 🛠️ Built with Modern Tools
- **Bubble Tea**: Interactive terminal UI
- **Cobra**: Powerful CLI framework
- **Viper**: Configuration management
- **pdfcpu**: PDF text extraction

## Installation

```bash
# Build the application
./build.sh

# Or manually with correct GOROOT
GOROOT=/home/linuxbrew/.linuxbrew/Cellar/go/1.24.5/libexec go build -o gengo
```

## Usage

### Basic Commands
```bash
# Show version
./gengo version

# Show help
./gengo --help

# Start interactive Bubble Tea mode
./gengo
```

### PDF Text Extraction
```bash
# Extract all text from PDF to stdout
./gengo pdf extract document.pdf

# Extract text to file
./gengo pdf extract document.pdf --output extracted.txt

# Extract specific pages
./gengo pdf extract document.pdf --pages 1,3,5

# Extract and clean text
./gengo pdf extract document.pdf --clean

# Get PDF information
./gengo pdf info document.pdf
```

### Interactive Mode
When you run `./gengo` without arguments, you enter the beautiful Bubble Tea interactive mode:

- Real-time typing with visual cursor
- Command history
- Available commands: `/exit`

## API Usage

The PDF extraction functionality can be used programmatically:

```go
import pdf "maai.solutions/gengo/internal"

// Create extractor
extractor := pdf.NewTextExtractor()

// Extract from file
text, err := extractor.ExtractFromFile("document.pdf")

// Extract from bytes
data, _ := os.ReadFile("document.pdf")
text, err := extractor.ExtractFromBytes(data)

// Extract to writer
var buf bytes.Buffer
err := extractor.ExtractFromFileToWriter("document.pdf", &buf)

// Extract specific pages
text, err := extractor.ExtractPages("document.pdf", []int{1, 3, 5})

// Clean extracted text
cleanText := extractor.CleanText(dirtyText)

// Get page count
count, err := extractor.GetPageCount("document.pdf")
```

## Project Structure

```
gengo/
├── cmd/                 # Cobra commands
│   ├── root.go         # Root command with Bubble Tea
│   ├── interactive.go  # Bubble Tea UI model
│   └── pdf.go          # PDF extraction commands
├── internal/           # Internal packages
│   ├── pdf_extractor.go # PDF text extraction logic
│   └── pdf_test.go     # Tests and examples
├── main.go             # Application entry point
├── go.mod              # Dependencies
└── build.sh            # Build script
```