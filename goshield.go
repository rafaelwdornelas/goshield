// GoShield - Advanced Go Source Code Obfuscator
// Copyright (c) 2024 - MIT License
//
// A powerful tool to protect your Go source code through multi-layer obfuscation:
// - Identifier renaming with Unicode lookalikes
// - String literal encryption
// - Integer transformation
// - JavaScript/embedded code obfuscation
// - Import aliasing
// - Comment removal
//
// Usage:
//   goshield -i input.go -o output.go [options]
//
// Options:
//   -i              Input file path (required)
//   -o              Output file path (required)
//   -seed           Seed for reproducible output
//   -no-ints        Disable integer obfuscation
//   -no-strings     Disable string obfuscation
//   -no-vars        Disable variable name obfuscation
//   -no-functions   Disable function name obfuscation
//   -no-imports     Disable import alias obfuscation
//   -v              Verbose output

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"hash/fnv"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// CONFIGURATION
// =============================================================================

var (
	inputFile  = flag.String("i", "", "Input Go file path")
	outputFile = flag.String("o", "", "Output Go file path")
	seed       = flag.String("seed", "", "Seed for reproducible obfuscation")
	verbose    = flag.Bool("v", false, "Verbose output")

	noInts      = flag.Bool("no-ints", false, "Disable integer obfuscation")
	noStrings   = flag.Bool("no-strings", false, "Disable string obfuscation")
	noVars      = flag.Bool("no-vars", false, "Disable variable obfuscation")
	noFunctions = flag.Bool("no-functions", false, "Disable function obfuscation")
	noImports   = flag.Bool("no-imports", false, "Disable import obfuscation")
)

// =============================================================================
// GLOBAL STATE
// =============================================================================

var nameMap = make(map[string]string)
var structTypeMapping = make(map[string]string)
var typeAliasMapping = make(map[string]string)

// Unicode lookalike characters for maximum confusion
var obfuscationChars = []rune{
	'O', '0', 'o', // O, zero, lowercase o
	'l', 'I', '1', // lowercase L, uppercase i, one
	'a', 'а', // Latin a, Cyrillic а
	'e', 'е', // Latin e, Cyrillic е
	'p', 'р', // Latin p, Cyrillic р
	'c', 'с', // Latin c, Cyrillic с
	'x', 'х', // Latin x, Cyrillic х
	'y', 'у', // Latin y, Cyrillic у
	'k', 'к', // Latin k, Cyrillic к
	'B', 'В', // Latin B, Cyrillic В
	'H', 'Н', // Latin H, Cyrillic Н
	'M', 'М', // Latin M, Cyrillic М
	'T', 'Т', // Latin T, Cyrillic Т
}

// Reserved names that should never be obfuscated (stdlib interfaces/methods)
var reservedNames = map[string]bool{
	"Error": true, "String": true,
	"Read": true, "Write": true, "Close": true, "Seek": true,
	"Len": true, "Cap": true, "Copy": true, "Append": true,
	"ServeHTTP": true, "Header": true, "Body": true, "Status": true, "StatusCode": true,
	"MarshalJSON": true, "UnmarshalJSON": true,
	"Context": true, "Err": true, "Done": true, "Value": true, "Deadline": true,
	"Lock": true, "Unlock": true, "RLock": true, "RUnlock": true,
	"ID": true, "URL": true, "URI": true, "HTML": true,
}

// =============================================================================
// LOGGING
// =============================================================================

func logDebug(format string, args ...interface{}) {
	if *verbose {
		fmt.Printf("  [DEBUG] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	fmt.Printf("  [+] "+format+"\n", args...)
}

func logError(format string, args ...interface{}) {
	fmt.Printf("  [!] "+format+"\n", args...)
}

func logSuccess(format string, args ...interface{}) {
	fmt.Printf("  [✓] "+format+"\n", args...)
}

// =============================================================================
// NAME GENERATION
// =============================================================================

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func generateObfuscatedName(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, length)
	result[0] = letters[rand.Intn(len(letters))]
	for i := 1; i < length; i++ {
		result[i] = obfuscationChars[rand.Intn(len(obfuscationChars))]
	}
	return string(result)
}

func getObfuscatedName(original string) string {
	if existing, ok := nameMap[original]; ok {
		return existing
	}

	var newName string
	for {
		newName = generateObfuscatedName(20)
		exists := false
		for _, v := range nameMap {
			if v == newName {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
	}

	nameMap[original] = newName
	logDebug("Rename: %s -> %s", original, newName)
	return newName
}

// =============================================================================
// STRING OBFUSCATION
// =============================================================================

func obfuscateStringLiteral(s string) string {
	if s == "" {
		return `""`
	}

	var parts []string
	for _, r := range s {
		switch rand.Intn(4) {
		case 0:
			parts = append(parts, fmt.Sprintf("string(%d)", r))
		case 1:
			parts = append(parts, fmt.Sprintf("string(0x%x)", r))
		case 2:
			offset := rand.Intn(50) + 1
			parts = append(parts, fmt.Sprintf("string(%d+%d)", int(r)-offset, offset))
		default:
			if r == '"' || r == '\\' || r > 127 {
				parts = append(parts, fmt.Sprintf("string(%d)", r))
			} else if r >= 32 && r < 127 {
				parts = append(parts, fmt.Sprintf(`"%c"`, r))
			} else {
				parts = append(parts, fmt.Sprintf("string(%d)", r))
			}
		}
	}
	return "(" + strings.Join(parts, "+") + ")"
}

func obfuscateFormatString(s string) string {
	// Regex to match Go format specifiers: %d, %s, %v, %f, %10.2f, %-5s, %+d, %#x, %%, etc.
	formatRe := regexp.MustCompile(`%[-+#0 ]*[0-9]*(\.[0-9]+)?[dsvftxXboqpeEgGUcTw%]`)

	// Find all format specifiers and their positions
	matches := formatRe.FindAllStringIndex(s, -1)
	if len(matches) == 0 {
		return obfuscateStringLiteral(s)
	}

	var parts []string
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Obfuscate text before this format specifier
		if start > lastEnd {
			textPart := s[lastEnd:start]
			if len(textPart) > 0 {
				parts = append(parts, obfuscateStringLiteral(textPart))
			}
		}

		// Keep format specifier as-is (quoted)
		formatSpec := s[start:end]
		parts = append(parts, fmt.Sprintf(`"%s"`, formatSpec))

		lastEnd = end
	}

	// Obfuscate remaining text after last format specifier
	if lastEnd < len(s) {
		textPart := s[lastEnd:]
		if len(textPart) > 0 {
			parts = append(parts, obfuscateStringLiteral(textPart))
		}
	}

	return "(" + strings.Join(parts, "+") + ")"
}

// =============================================================================
// INTEGER OBFUSCATION
// =============================================================================

func obfuscateInteger(n int64) string {
	switch rand.Intn(4) {
	case 0:
		x := rand.Int63n(1000) + 1
		return fmt.Sprintf("(%d+%d)", n-x, x)
	case 1:
		x := rand.Int63n(1000) + 1
		return fmt.Sprintf("(%d-%d)", n+x, x)
	case 2:
		x := rand.Int63n(1000) + 1
		return fmt.Sprintf("(%d^%d)", n^x, x)
	default:
		x := rand.Int63n(10) + 2
		return fmt.Sprintf("(%d/%d)", n*x, x)
	}
}

// =============================================================================
// AST UTILITIES
// =============================================================================

func writeAST(filename string, file *ast.File, fset *token.FileSet) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	return cfg.Fprint(f, fset, file)
}

func parseFile(filename string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0) // No comments
	if err != nil {
		return nil, nil, err
	}
	return file, fset, nil
}

// =============================================================================
// OBFUSCATOR STRUCT
// =============================================================================

type Obfuscator struct {
	file            *ast.File
	fset            *token.FileSet
	declaredFuncs   map[string]bool
	declaredMethods map[string]bool
	importAliases   map[string]string
	structFields    map[string]bool
	typeNames       map[string]bool
	structTypes     map[string]bool
	fieldNames      map[string]string
}

func NewObfuscator(file *ast.File, fset *token.FileSet) *Obfuscator {
	return &Obfuscator{
		file:            file,
		fset:            fset,
		declaredFuncs:   make(map[string]bool),
		declaredMethods: make(map[string]bool),
		importAliases:   make(map[string]string),
		structFields:    make(map[string]bool),
		typeNames:       make(map[string]bool),
		structTypes:     make(map[string]bool),
		fieldNames:      make(map[string]string),
	}
}

// =============================================================================
// COLLECTION PASSES
// =============================================================================

func (o *Obfuscator) collectTypeNames() {
	ast.Inspect(o.file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			o.typeNames[typeSpec.Name.Name] = true
		}
		return true
	})
}

func (o *Obfuscator) collectDeclaredFunctions() {
	ast.Inspect(o.file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		name := fn.Name.Name
		if name == "main" || name == "init" {
			return true
		}
		if fn.Recv == nil {
			o.declaredFuncs[name] = true
		} else {
			o.declaredMethods[name] = true
		}
		return true
	})
}

func (o *Obfuscator) collectStructFields() {
	ast.Inspect(o.file, func(n ast.Node) bool {
		structType, ok := n.(*ast.StructType)
		if !ok || structType.Fields == nil {
			return true
		}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if !reservedNames[name.Name] {
					o.structFields[name.Name] = true
				}
			}
		}
		return true
	})
}

func (o *Obfuscator) collectStructTypes() {
	ast.Inspect(o.file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
			originalName := typeSpec.Name.Name
			obfuscatedName := getObfuscatedName(originalName)
			o.structTypes[originalName] = true
			structTypeMapping[originalName] = obfuscatedName
		} else {
			originalName := typeSpec.Name.Name
			obfuscatedName := getObfuscatedName(originalName)
			typeAliasMapping[originalName] = obfuscatedName
		}
		return true
	})
}

// =============================================================================
// OBFUSCATION PASSES
// =============================================================================

func (o *Obfuscator) obfuscateConsts() {
	ast.Inspect(o.file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if ok && genDecl.Tok == token.CONST {
			genDecl.Tok = token.VAR
		}
		return true
	})
}

func (o *Obfuscator) obfuscateImports() {
	if *noImports {
		return
	}
	for _, decl := range o.file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			path := strings.Trim(importSpec.Path.Value, `"`)
			parts := strings.Split(path, "/")
			baseName := parts[len(parts)-1]
			alias := getObfuscatedName(baseName)
			o.importAliases[baseName] = alias
			importSpec.Name = &ast.Ident{Name: alias, NamePos: importSpec.Path.Pos()}
		}
	}
}

func (o *Obfuscator) updateImportReferences() {
	if *noImports {
		return
	}
	ast.Inspect(o.file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok := sel.X.(*ast.Ident); ok {
			if alias, exists := o.importAliases[ident.Name]; exists {
				ident.Name = alias
			}
		}
		return true
	})
}

func (o *Obfuscator) obfuscateStructTypes() {
	fieldNameSet := make(map[string]bool)
	ast.Inspect(o.file, func(n ast.Node) bool {
		structType, ok := n.(*ast.StructType)
		if !ok || structType.Fields == nil {
			return true
		}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				fieldNameSet[name.Name] = true
			}
		}
		return true
	})

	ast.Inspect(o.file, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if fieldNameSet[ident.Name] {
			return true
		}
		if obfuscated, exists := structTypeMapping[ident.Name]; exists {
			ident.Name = obfuscated
		}
		if obfuscated, exists := typeAliasMapping[ident.Name]; exists {
			ident.Name = obfuscated
		}
		return true
	})
}

func (o *Obfuscator) obfuscateVariables() {
	if *noVars {
		return
	}

	packageVars := make(map[string]bool)
	for _, decl := range o.file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range valueSpec.Names {
				if name.Name != "_" && !reservedNames[name.Name] {
					packageVars[name.Name] = true
				}
			}
		}
	}

	ast.Inspect(o.file, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if reservedNames[ident.Name] || o.structTypes[ident.Name] || o.structFields[ident.Name] {
			return true
		}
		if _, isTypeAlias := typeAliasMapping[ident.Name]; isTypeAlias {
			return true
		}
		if packageVars[ident.Name] {
			ident.Name = getObfuscatedName(ident.Name)
			return true
		}
		if ident.Obj != nil && ident.Obj.Kind == ast.Var && ident.Name != "_" {
			ident.Name = getObfuscatedName(ident.Name)
		}
		return true
	})
}

func (o *Obfuscator) obfuscateFunctions() {
	if *noFunctions {
		return
	}

	ast.Inspect(o.file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		name := fn.Name.Name
		if o.declaredFuncs[name] || o.declaredMethods[name] {
			fn.Name.Name = getObfuscatedName(name)
		}
		return true
	})

	ast.Inspect(o.file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if o.declaredFuncs[ident.Name] {
				ident.Name = getObfuscatedName(ident.Name)
			}
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if o.declaredMethods[sel.Sel.Name] {
				sel.Sel.Name = getObfuscatedName(sel.Sel.Name)
			}
		}
		return true
	})

	ast.Inspect(o.file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if o.declaredMethods[sel.Sel.Name] && !o.structFields[sel.Sel.Name] {
			sel.Sel.Name = getObfuscatedName(sel.Sel.Name)
		}
		return true
	})
}

// =============================================================================
// TEXT-BASED OBFUSCATION
// =============================================================================

func obfuscateBacktickStrings(content string) string {
	if *noStrings {
		return content
	}

	re := regexp.MustCompile("(?s)`[^`]+`")
	count := 0

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		innerContent := match[1 : len(match)-1]

		if len(innerContent) < 20 {
			return match
		}

		// Skip struct tags
		if strings.Contains(innerContent, "json:") || strings.Contains(innerContent, "xml:") ||
			strings.Contains(innerContent, "yaml:") || strings.Contains(innerContent, "gorm:") {
			return match
		}

		// Check if it looks like code (JavaScript, SQL, etc.)
		isCode := strings.Contains(innerContent, "function") ||
			strings.Contains(innerContent, "await") ||
			strings.Contains(innerContent, "async") ||
			strings.Contains(innerContent, "const ") ||
			strings.Contains(innerContent, "var ") ||
			strings.Contains(innerContent, "let ") ||
			strings.Contains(innerContent, "try {") ||
			strings.Contains(innerContent, "catch") ||
			strings.Contains(innerContent, "return ") ||
			strings.Contains(innerContent, "SELECT ") ||
			strings.Contains(innerContent, "INSERT ") ||
			strings.Contains(innerContent, "UPDATE ")

		if !isCode {
			return match
		}

		var parts []string
		for i := 0; i < len(innerContent); i++ {
			c := innerContent[i]
			switch rand.Intn(3) {
			case 0:
				parts = append(parts, fmt.Sprintf("string(%d)", c))
			case 1:
				parts = append(parts, fmt.Sprintf("string(0x%x)", c))
			default:
				if c >= 32 && c < 127 && c != '"' && c != '\\' && c != '\'' {
					parts = append(parts, fmt.Sprintf(`"%c"`, c))
				} else {
					parts = append(parts, fmt.Sprintf("string(%d)", c))
				}
			}
		}

		count++
		return "(" + strings.Join(parts, "+") + ")"
	})

	if count > 0 {
		logInfo("Embedded code strings: %d", count)
	}
	return result
}

func obfuscateStringsInText(content string) string {
	if *noStrings {
		return content
	}

	lines := strings.Split(content, "\n")
	inImportBlock := false
	count := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}
		if inImportBlock || strings.HasPrefix(trimmed, "import ") {
			continue
		}

		if strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "case ") ||
			strings.HasPrefix(trimmed, "Set(") ||
			strings.HasPrefix(trimmed, ".Set(") ||
			strings.Contains(line, ".Set(") ||
			strings.Contains(line, "`json:") ||
			strings.Contains(line, "`xml:") ||
			strings.Contains(line, "`yaml:") ||
			strings.Contains(line, "Flag") ||
			strings.Contains(line, "flag.") ||
			strings.Contains(line, "launcher.") {
			continue
		}

		typeAliasPattern := regexp.MustCompile(`^\s*[^\s]+\s+[^\s]+\s*=\s*"[^"]*"\s*$`)
		if typeAliasPattern.MatchString(line) {
			continue
		}

		isVarAssignment := strings.Contains(line, "=") &&
			!strings.Contains(line, "==") &&
			!strings.Contains(line, "!=") &&
			!strings.Contains(line, "(")

		re := regexp.MustCompile(`"([^"\\]|\\.)*"`)

		lines[i] = re.ReplaceAllStringFunc(line, func(match string) string {
			s, err := strconv.Unquote(match)
			if err != nil {
				return match
			}
			if len(s) < 3 {
				return match
			}
			if strings.Contains(s, "\\") {
				return match
			}
			if strings.Contains(s, "://") && !isVarAssignment {
				return match
			}
			count++
			if strings.Contains(s, "%") {
				return obfuscateFormatString(s)
			}
			return obfuscateStringLiteral(s)
		})
	}

	logInfo("String literals: %d", count)
	return strings.Join(lines, "\n")
}

func obfuscateIntegersInText(content string) string {
	if *noInts {
		return content
	}

	lines := strings.Split(content, "\n")
	count := 0

	for i, line := range lines {
		if strings.Contains(line, "string(") ||
			strings.Contains(line, `"`) ||
			strings.Contains(line, "`") ||
			strings.Contains(line, "'") {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") ||
			strings.Contains(line, "func (") ||
			strings.HasPrefix(trimmed, "import") ||
			strings.HasPrefix(trimmed, "package") ||
			strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "var ") {
			continue
		}

		re := regexp.MustCompile(`([\s(=:,])(\d+)([\s),;\]])`)

		lines[i] = re.ReplaceAllStringFunc(line, func(match string) string {
			numRe := regexp.MustCompile(`\d+`)
			numStr := numRe.FindString(match)
			n, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil || n <= 10 || n > 100000 {
				return match
			}
			count++
			prefix := match[0:1]
			suffix := match[len(match)-1:]
			return prefix + obfuscateInteger(n) + suffix
		})
	}

	if count > 0 {
		logInfo("Integer literals: %d", count)
	}
	return strings.Join(lines, "\n")
}

// =============================================================================
// MAIN
// =============================================================================

func printBanner() {
	fmt.Println(`
   ██████╗  ██████╗ ███████╗██╗  ██╗██╗███████╗██╗     ██████╗
  ██╔════╝ ██╔═══██╗██╔════╝██║  ██║██║██╔════╝██║     ██╔══██╗
  ██║  ███╗██║   ██║███████╗███████║██║█████╗  ██║     ██║  ██║
  ██║   ██║██║   ██║╚════██║██╔══██║██║██╔══╝  ██║     ██║  ██║
  ╚██████╔╝╚██████╔╝███████║██║  ██║██║███████╗███████╗██████╔╝
   ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚═════╝
                    Go Source Code Obfuscator v1.0
`)
}

func main() {
	flag.Parse()

	printBanner()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Usage: goshield -i <input.go> -o <output.go> [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *seed != "" {
		rand.Seed(int64(hashString(*seed)))
		logInfo("Using seed: %s", *seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	fmt.Printf("\n  Input:  %s\n", *inputFile)
	fmt.Printf("  Output: %s\n\n", *outputFile)

	// Parse
	file, fset, err := parseFile(*inputFile)
	if err != nil {
		logError("Parse failed: %v", err)
		os.Exit(1)
	}

	// Collect
	obf := NewObfuscator(file, fset)
	obf.collectTypeNames()
	obf.collectDeclaredFunctions()
	obf.collectStructFields()
	obf.collectStructTypes()

	fmt.Println("  Processing...")

	// AST obfuscation
	obf.obfuscateConsts()
	obf.obfuscateImports()
	obf.updateImportReferences()
	obf.obfuscateStructTypes()
	obf.obfuscateVariables()
	obf.obfuscateFunctions()

	// Write intermediate
	if err := writeAST(*outputFile, file, fset); err != nil {
		logError("Write failed: %v", err)
		os.Exit(1)
	}

	// Text obfuscation
	content, err := ioutil.ReadFile(*outputFile)
	if err != nil {
		logError("Read failed: %v", err)
		os.Exit(1)
	}

	text := string(content)
	text = obfuscateBacktickStrings(text)
	text = obfuscateStringsInText(text)
	text = obfuscateIntegersInText(text)

	if err := ioutil.WriteFile(*outputFile, []byte(text), 0644); err != nil {
		logError("Final write failed: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	logSuccess("Obfuscation complete!")
	logSuccess("Identifiers renamed: %d", len(nameMap))
	fmt.Printf("\n  Output saved to: %s\n\n", *outputFile)
}
