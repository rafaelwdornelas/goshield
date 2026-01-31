# GoShield 🛡️

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Platform-Cross--Platform-blue?style=for-the-badge" alt="Platform">
</p>

<p align="center">
  <b>Advanced Go Source Code Obfuscator</b><br>
  Protect your Go applications with multi-layer obfuscation
</p>

---

## ✨ Features

GoShield provides comprehensive protection for your Go source code through multiple obfuscation techniques:

| Feature | Description |
|---------|-------------|
| 🔤 **Identifier Renaming** | Renames variables, functions, methods, and types using Unicode lookalikes (Cyrillic/Latin mix) |
| 📝 **String Encryption** | Converts string literals to runtime-computed expressions |
| 🔢 **Integer Obfuscation** | Transforms numeric literals using mathematical expressions |
| 📦 **Import Aliasing** | Adds random aliases to all imports |
| 💻 **Embedded Code** | Obfuscates JavaScript, SQL, and other embedded code in backtick strings |
| 🗑️ **Comment Removal** | Automatically strips all comments from the output |
| 🏗️ **Type Obfuscation** | Renames struct types and type aliases |
| 📦 **Minification** | Removes empty lines and reduces code to compact form (~65% line reduction) |

## 🚀 Installation

```bash
# Clone the repository
git clone https://github.com/rafaelwdornelas/goshield.git
cd goshield

# Build
go build -o goshield goshield.go

# Or install directly
go install github.com/rafaelwdornelas/goshield@latest
```

## 📖 Usage

### Basic Usage

```bash
goshield -i input.go -o output.go
```

### With Minification

```bash
goshield -i main.go -o obfuscated.go -minify
```

### With All Options

```bash
goshield -i main.go -o obfuscated.go -minify -seed mysecret -v
```

### All Options

| Flag | Description | Default |
|------|-------------|---------|
| `-i` | Input Go file path | (required) |
| `-o` | Output Go file path | (required) |
| `-seed` | Seed for reproducible obfuscation | random |
| `-minify` | Minify output (remove empty lines, compact code) | false |
| `-v` | Verbose output | false |
| `-no-strings` | Disable string obfuscation | false |
| `-no-ints` | Disable integer obfuscation | false |
| `-no-vars` | Disable variable obfuscation | false |
| `-no-functions` | Disable function obfuscation | false |
| `-no-imports` | Disable import obfuscation | false |

## 📋 Example

### Before (input.go)

```go
package main

import "fmt"

func main() {
    message := "Hello, World!"
    count := 42
    fmt.Println(message, count)
}
```

### After (output.go)

```go
package main

import Bа1Тxk0МоHpрT "fmt"

func main() {
    xМНlТ0аeуВkрТpО := (string(72)+string(0x65)+string(108)+string(0x6c)+string(111)+string(44)+string(0x20)+string(87)+string(111)+string(114)+string(0x6c)+string(100)+string(33))
    kТ0рВНМxаpОеl := (18+24)
    Bа1Тxk0МоHpрT.Println(xМНlТ0аeуВkрТpО, kТ0рВНМxаpОеl)
}
```

### After with Minification (-minify)

```go
package main
import Bа1Тxk0МоHpрT "fmt"
func main() { xМНlТ0аeуВkрТpО := (string(72)+string(0x65)+string(108)+string(0x6c)+string(111)+string(44)+string(0x20)+string(87)+string(111)+string(114)+string(0x6c)+string(100)+string(33)); kТ0рВНМxаpОеl := (18+24); Bа1Тxk0МоHpрT.Println(xМНlТ0аeуВkрТpО, kТ0рВНМxаpОеl) }
```

## 🔒 What Gets Obfuscated

### ✅ Obfuscated
- Local and package-level variables
- Function and method names
- Struct type names
- Type aliases
- Import aliases
- String literals (converted to character code concatenations)
- Integer literals (converted to mathematical expressions)
- Embedded JavaScript/SQL in backtick strings

### ⚠️ Preserved (for compatibility)
- Struct field names (required for JSON/GOB/XML serialization)
- Reserved interface methods (`Error`, `String`, `Read`, `Write`, etc.)
- `main` and `init` functions
- Struct tags (json, xml, yaml, gorm)

## 🎯 Use Cases

- **Protect proprietary algorithms** - Make reverse engineering significantly harder
- **Distribute closed-source binaries** - Ship obfuscated source to clients
- **License protection** - Complicate unauthorized modifications
- **Security through obscurity** - Add an extra layer of protection

## ⚠️ Important Notes

1. **Backup your code** - Always keep the original source code safe
2. **Test thoroughly** - Verify the obfuscated code works correctly
3. **Reproducible builds** - Use `-seed` flag for consistent output
4. **Single file** - Currently processes one file at a time

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by various obfuscation techniques from the security community
- Built with Go's powerful AST package

---

<p align="center">
  Made with ❤️ for the Go community
</p>
