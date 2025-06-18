package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"relay/pkg/parser"
	"relay/pkg/runtime"
)

const PROMPT = "relay> "
const MULTILINE_PROMPT = "    | "

// REPL represents the Read-Eval-Print Loop
type REPL struct {
	scanner   *bufio.Scanner
	output    io.Writer
	evaluator *runtime.Evaluator
	showAST   bool // Toggle between AST mode and execution mode
}

// New creates a new REPL instance
func New(input io.Reader, output io.Writer) *REPL {
	return &REPL{
		scanner:   bufio.NewScanner(input),
		output:    output,
		evaluator: runtime.NewEvaluator(),
		showAST:   false, // Default to execution mode
	}
}

// Start begins the REPL session
func (r *REPL) Start() {
	fmt.Fprint(r.output, welcome())

	for {
		fmt.Fprint(r.output, PROMPT)

		if !r.scanner.Scan() {
			break
		}

		line := strings.TrimSpace(r.scanner.Text())

		// Handle empty input
		if line == "" {
			continue
		}

		// Handle special commands
		if strings.HasPrefix(line, ":") {
			r.handleCommand(line)
			continue
		}

		// Check if we need multiple lines (incomplete expression)
		input := line
		for r.needsMoreInput(input) {
			fmt.Fprint(r.output, MULTILINE_PROMPT)
			if !r.scanner.Scan() {
				break
			}
			nextLine := strings.TrimSpace(r.scanner.Text())
			if nextLine == "" {
				break
			}
			input += "\n" + nextLine
		}

		// Parse and evaluate the input
		r.evaluate(input)
	}

	fmt.Fprintln(r.output, "\nGoodbye!")
}

// needsMoreInput checks if the input is incomplete and needs more lines
func (r *REPL) needsMoreInput(input string) bool {
	// Simple heuristic: check for unmatched braces
	openBraces := strings.Count(input, "{")
	closeBraces := strings.Count(input, "}")

	openParens := strings.Count(input, "(")
	closeParens := strings.Count(input, ")")

	openBrackets := strings.Count(input, "[")
	closeBrackets := strings.Count(input, "]")

	return openBraces > closeBraces || openParens > closeParens || openBrackets > closeBrackets
}

// evaluate parses and evaluates the input
func (r *REPL) evaluate(input string) {
	// Parse the input
	program, err := parser.Parse("repl", strings.NewReader(input))
	if err != nil {
		fmt.Fprintf(r.output, "Parse error: %v\n", err)
		return
	}

	if len(program.Expressions) == 0 {
		fmt.Fprintln(r.output, "Empty program")
		return
	}

	// Show AST if in AST mode
	if r.showAST {
		r.printAST(program)
		return
	}

	// Execute the expressions
	for _, expr := range program.Expressions {
		result, err := r.evaluator.Evaluate(expr)
		if err != nil {
			fmt.Fprintf(r.output, "Runtime error: %v\n", err)
			continue
		}

		// Only print results for expressions that aren't assignments
		// (assignments already show their value)
		if expr.SetExpr == nil {
			// Show the result with some formatting
			if result.Type == runtime.ValueTypeNil {
				fmt.Fprintln(r.output, "nil")
			} else {
				fmt.Fprintln(r.output, result.String())
			}
		} else {
			// For set expressions, show: variable = value
			fmt.Fprintf(r.output, "%s = %s\n", expr.SetExpr.Variable, result.String())
		}
	}
}

// printAST prints a formatted representation of the AST
func (r *REPL) printAST(program *parser.Program) {
	if len(program.Expressions) == 0 {
		fmt.Fprintln(r.output, "Empty program")
		return
	}

	fmt.Fprintln(r.output, "=== Complete AST Tree ===")
	fmt.Fprintln(r.output, "Program")
	for i, expr := range program.Expressions {
		fmt.Fprintf(r.output, "├─ [%d] %s\n", i, r.formatExpressionRecursive(expr, "│  "))
	}
	fmt.Fprintln(r.output, "=========================")
}

// formatExpressionRecursive returns a detailed recursive representation of an expression
func (r *REPL) formatExpressionRecursive(expr *parser.Expression, indent string) string {
	result := ""

	switch {
	case expr.StructExpr != nil:
		s := expr.StructExpr
		result = fmt.Sprintf("StructExpr: '%s'\n", s.Name)
		for i, field := range s.Fields {
			prefix := "├─ "
			childIndent := indent + "│  "
			if i == len(s.Fields)-1 {
				prefix = "└─ "
				childIndent = indent + "   "
			}
			result += fmt.Sprintf("%s%sField: '%s'\n", indent, prefix, field.Name)
			result += fmt.Sprintf("%s   ├─ Type: %s\n", indent, r.formatTypeRefRecursive(field.Type, childIndent+"   "))
		}

	case expr.ProtocolExpr != nil:
		p := expr.ProtocolExpr
		result = fmt.Sprintf("ProtocolExpr: '%s'\n", p.Name)
		for i, method := range p.Methods {
			prefix := "├─ "
			childIndent := indent + "│  "
			if i == len(p.Methods)-1 {
				prefix = "└─ "
				childIndent = indent + "   "
			}
			result += fmt.Sprintf("%s%sMethod: '%s'\n", indent, prefix, method.Name)
			if len(method.Parameters) > 0 {
				result += fmt.Sprintf("%s   ├─ Parameters:\n", indent)
				for j, param := range method.Parameters {
					paramPrefix := "├─ "
					if j == len(method.Parameters)-1 {
						paramPrefix = "└─ "
					}
					result += fmt.Sprintf("%s   │  %s%s: %s\n", indent, paramPrefix, param.Name, r.formatTypeRefRecursive(param.Type, childIndent+"   │  "))
				}
			}
			if method.ReturnType != nil {
				result += fmt.Sprintf("%s   └─ ReturnType: %s\n", indent, r.formatTypeRefRecursive(method.ReturnType, childIndent+"   "))
			}
		}

	case expr.ServerExpr != nil:
		s := expr.ServerExpr
		result = fmt.Sprintf("ServerExpr: '%s'\n", s.Name)
		if s.Protocol != nil {
			result += fmt.Sprintf("%s├─ Implements: '%s'\n", indent, *s.Protocol)
		}
		if s.Body != nil {
			result += fmt.Sprintf("%s└─ Body:\n", indent)
			result += r.formatServerBodyRecursive(s.Body, indent+"   ")
		}

	case expr.FunctionExpr != nil:
		f := expr.FunctionExpr
		name := "anonymous"
		if f.Name != nil {
			name = *f.Name
		}
		result = fmt.Sprintf("FunctionExpr: '%s'\n", name)
		if len(f.Parameters) > 0 {
			result += fmt.Sprintf("%s├─ Parameters:\n", indent)
			for i, param := range f.Parameters {
				prefix := "├─ "
				if i == len(f.Parameters)-1 {
					prefix = "└─ "
				}
				result += fmt.Sprintf("%s│  %s%s: %s\n", indent, prefix, param.Name, r.formatTypeRefRecursive(param.Type, indent+"│  "))
			}
		}
		if f.ReturnType != nil {
			result += fmt.Sprintf("%s├─ ReturnType: %s\n", indent, r.formatTypeRefRecursive(f.ReturnType, indent+"│  "))
		}
		if f.Body != nil {
			result += fmt.Sprintf("%s└─ Body:\n", indent)
			result += r.formatBlockRecursive(f.Body, indent+"   ")
		}

	case expr.TemplateExpr != nil:
		t := expr.TemplateExpr
		result = fmt.Sprintf("TemplateExpr: '%s'\n", t.Path)
		if t.FromFunc != nil {
			result += fmt.Sprintf("%s└─ FromFunc: '%s'\n", indent, t.FromFunc.Name)
			if len(t.FromFunc.Parameters) > 0 {
				result += fmt.Sprintf("%s   └─ Parameters:\n", indent)
				for i, param := range t.FromFunc.Parameters {
					prefix := "├─ "
					if i == len(t.FromFunc.Parameters)-1 {
						prefix = "└─ "
					}
					result += fmt.Sprintf("%s      %s%s: %s\n", indent, prefix, param.Name, r.formatTypeRefRecursive(param.Type, indent+"      "))
				}
			}
		}

	case expr.ConfigExpr != nil:
		c := expr.ConfigExpr
		result = fmt.Sprintf("ConfigExpr:\n")
		for i, field := range c.Fields {
			prefix := "├─ "
			childIndent := indent + "│  "
			if i == len(c.Fields)-1 {
				prefix = "└─ "
				childIndent = indent + "   "
			}
			result += fmt.Sprintf("%s%sField: '%s'\n", indent, prefix, field.Key)
			result += fmt.Sprintf("%s   └─ Value: %s\n", indent, r.formatExpressionRecursive(field.Value, childIndent+"   "))
		}

	case expr.SetExpr != nil:
		s := expr.SetExpr
		result = fmt.Sprintf("SetExpr: '%s'\n", s.Variable)
		result += fmt.Sprintf("%s└─ Value: %s\n", indent, r.formatExpressionRecursive(s.Value, indent+"   "))

	case expr.ReturnExpr != nil:
		ret := expr.ReturnExpr
		result = fmt.Sprintf("ReturnExpr:\n")
		if ret.Value != nil {
			result += fmt.Sprintf("%s└─ Value: %s\n", indent, r.formatExpressionRecursive(ret.Value, indent+"   "))
		}

	case expr.IfExpr != nil:
		i := expr.IfExpr
		result = fmt.Sprintf("IfExpr:\n")
		result += fmt.Sprintf("%s├─ Condition: %s\n", indent, r.formatExpressionRecursive(i.Condition, indent+"│  "))
		result += fmt.Sprintf("%s├─ ThenBlock:\n", indent)
		result += r.formatBlockRecursive(i.ThenBlock, indent+"│  ")
		if i.ElseBlock != nil {
			result += fmt.Sprintf("%s└─ ElseBlock:\n", indent)
			result += r.formatBlockRecursive(i.ElseBlock, indent+"   ")
		}

	case expr.ForExpr != nil:
		f := expr.ForExpr
		result = fmt.Sprintf("ForExpr: '%s'\n", f.Variable)
		result += fmt.Sprintf("%s├─ Collection: %s\n", indent, r.formatExpressionRecursive(f.Collection, indent+"│  "))
		result += fmt.Sprintf("%s└─ Body:\n", indent)
		result += r.formatBlockRecursive(f.Body, indent+"   ")

	case expr.TryExpr != nil:
		t := expr.TryExpr
		result = fmt.Sprintf("TryExpr:\n")
		result += fmt.Sprintf("%s├─ TryBlock:\n", indent)
		result += r.formatBlockRecursive(t.TryBlock, indent+"│  ")
		if t.CatchVar != nil {
			result += fmt.Sprintf("%s├─ CatchVar: '%s'\n", indent, *t.CatchVar)
		}
		result += fmt.Sprintf("%s└─ CatchBlock:\n", indent)
		result += r.formatBlockRecursive(t.CatchBlock, indent+"   ")

	case expr.DispatchExpr != nil:
		d := expr.DispatchExpr
		result = fmt.Sprintf("DispatchExpr:\n")
		result += fmt.Sprintf("%s├─ Value: %s\n", indent, r.formatExpressionRecursive(d.Value, indent+"│  "))
		result += fmt.Sprintf("%s└─ Cases:\n", indent)
		for i, c := range d.Cases {
			prefix := "├─ "
			childIndent := indent + "   │  "
			if i == len(d.Cases)-1 {
				prefix = "└─ "
				childIndent = indent + "      "
			}
			result += fmt.Sprintf("%s   %sCase:\n", indent, prefix)
			result += fmt.Sprintf("%s   │  ├─ Pattern: %s\n", indent, r.formatLiteralRecursive(c.Pattern, childIndent))
			result += fmt.Sprintf("%s   │  └─ Body: %s\n", indent, r.formatExpressionRecursive(c.Body, childIndent))
		}

	case expr.ThrowExpr != nil:
		t := expr.ThrowExpr
		result = fmt.Sprintf("ThrowExpr:\n")
		if t.Value != nil {
			result += fmt.Sprintf("%s└─ Value: %s\n", indent, r.formatExpressionRecursive(t.Value, indent+"   "))
		}

	case expr.Binary != nil:
		result = r.formatBinaryExprRecursive(expr.Binary, indent)

	default:
		result = "UnknownExpr"
	}

	return result
}

// formatTypeRefRecursive formats a type reference recursively
func (r *REPL) formatTypeRefRecursive(typeRef *parser.TypeRef, indent string) string {
	if typeRef == nil {
		return "nil"
	}

	switch {
	case typeRef.Array != nil:
		return fmt.Sprintf("Array[%s]", r.formatTypeRefRecursive(typeRef.Array, indent))
	case typeRef.Function != nil:
		f := typeRef.Function
		result := "Function("
		for i, param := range f.Parameters {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%s: %s", param.Name, r.formatTypeRefRecursive(param.Type, indent))
		}
		result += ")"
		if f.ReturnType != nil {
			result += fmt.Sprintf(" -> %s", r.formatTypeRefRecursive(f.ReturnType, indent))
		}
		return result
	case typeRef.Parameterized != nil:
		p := typeRef.Parameterized
		result := fmt.Sprintf("%s(", p.Name)
		for i, arg := range p.Args {
			if i > 0 {
				result += ", "
			}
			result += r.formatTypeRefRecursive(arg, indent)
		}
		result += ")"
		return result
	case typeRef.Name != "":
		return typeRef.Name
	default:
		return "UnknownType"
	}
}

// formatServerBodyRecursive formats a server body recursively
func (r *REPL) formatServerBodyRecursive(body *parser.ServerBody, indent string) string {
	result := ""
	for i, element := range body.Elements {
		prefix := "├─ "
		childIndent := indent + "│  "
		if i == len(body.Elements)-1 {
			prefix = "└─ "
			childIndent = indent + "   "
		}

		if element.State != nil {
			result += fmt.Sprintf("%s%sState:\n", indent, prefix)
			for j, field := range element.State.Fields {
				fieldPrefix := "├─ "
				if j == len(element.State.Fields)-1 {
					fieldPrefix = "└─ "
				}
				result += fmt.Sprintf("%s   %sField: '%s'\n", indent, fieldPrefix, field.Name)
				result += fmt.Sprintf("%s   │  ├─ Type: %s\n", indent, r.formatTypeRefRecursive(field.Type, childIndent+"   │  "))
				if field.DefaultValue != nil {
					result += fmt.Sprintf("%s   │  └─ Default: %s\n", indent, r.formatLiteralRecursive(field.DefaultValue, childIndent+"   │  "))
				}
			}
		}

		if element.Receive != nil {
			recv := element.Receive
			result += fmt.Sprintf("%s%sReceive: '%s'\n", indent, prefix, recv.Name)
			if len(recv.Parameters) > 0 {
				result += fmt.Sprintf("%s   ├─ Parameters:\n", indent)
				for j, param := range recv.Parameters {
					paramPrefix := "├─ "
					if j == len(recv.Parameters)-1 {
						paramPrefix = "└─ "
					}
					result += fmt.Sprintf("%s   │  %s%s: %s\n", indent, paramPrefix, param.Name, r.formatTypeRefRecursive(param.Type, childIndent+"   │  "))
				}
			}
			if recv.ReturnType != nil {
				result += fmt.Sprintf("%s   ├─ ReturnType: %s\n", indent, r.formatTypeRefRecursive(recv.ReturnType, childIndent+"   "))
			}
			if recv.Body != nil {
				result += fmt.Sprintf("%s   └─ Body:\n", indent)
				result += r.formatBlockRecursive(recv.Body, childIndent+"   ")
			}
		}
	}
	return result
}

// formatBlockRecursive formats a block recursively
func (r *REPL) formatBlockRecursive(block *parser.Block, indent string) string {
	result := ""
	for i, expr := range block.Expressions {
		prefix := "├─ "
		childIndent := indent + "│  "
		if i == len(block.Expressions)-1 {
			prefix = "└─ "
			childIndent = indent + "   "
		}
		result += fmt.Sprintf("%s%s%s", indent, prefix, r.formatExpressionRecursive(expr, childIndent))
	}
	return result
}

// formatLiteralRecursive formats a literal recursively
func (r *REPL) formatLiteralRecursive(lit *parser.Literal, indent string) string {
	if lit == nil {
		return "nil"
	}

	switch {
	case lit.String != nil:
		return fmt.Sprintf("String: %s", *lit.String)
	case lit.Number != nil:
		return fmt.Sprintf("Number: %v", *lit.Number)
	case lit.Bool != nil:
		return fmt.Sprintf("Bool: %s", *lit.Bool)
	case lit.Symbol != nil:
		return fmt.Sprintf("Symbol: %s", *lit.Symbol)
	case lit.Array != nil:
		result := "Array:\n"
		for i, elem := range lit.Array.Elements {
			prefix := "├─ "
			if i == len(lit.Array.Elements)-1 {
				prefix = "└─ "
			}
			result += fmt.Sprintf("%s%s%s\n", indent, prefix, r.formatLiteralRecursive(elem, indent+"   "))
		}
		return result
	case lit.FuncCall != nil:
		f := lit.FuncCall
		result := fmt.Sprintf("FuncCall: '%s'\n", f.Name)
		if len(f.Args) > 0 {
			result += fmt.Sprintf("%s└─ Args:\n", indent)
			for i, arg := range f.Args {
				prefix := "├─ "
				if i == len(f.Args)-1 {
					prefix = "└─ "
				}
				result += fmt.Sprintf("%s   %s%s\n", indent, prefix, r.formatLiteralRecursive(arg, indent+"      "))
			}
		}
		return result
	default:
		return "UnknownLiteral"
	}
}

// formatBinaryExprRecursive formats binary expressions recursively
func (r *REPL) formatBinaryExprRecursive(binary *parser.BinaryExpr, indent string) string {
	// If no operations, just return the left side
	if len(binary.Right) == 0 {
		return r.formatUnaryExprRecursive(binary.Left, indent)
	}

	// Build the expression string
	result := r.formatUnaryExprRecursive(binary.Left, indent)
	for _, op := range binary.Right {
		result += fmt.Sprintf(" %s %s", op.Op, r.formatUnaryExprRecursive(op.Right, indent))
	}

	return result
}

// formatUnaryExprRecursive formats unary expressions recursively
func (r *REPL) formatUnaryExprRecursive(expr *parser.UnaryExpr, indent string) string {
	result := ""
	if expr.Op != nil {
		result = fmt.Sprintf("UnaryExpr: %s ", *expr.Op)
	}
	result += r.formatPrimaryExprRecursive(expr.Primary, indent)
	return result
}

// formatPrimaryExprRecursive formats primary expressions recursively
func (r *REPL) formatPrimaryExprRecursive(expr *parser.PrimaryExpr, indent string) string {
	if len(expr.Access) == 0 {
		// No method calls, just return the base
		return r.formatBaseExprRecursive(expr.Base, indent)
	}

	// There are method calls, format as a proper chain
	result := fmt.Sprintf("MethodChain:\n")
	result += fmt.Sprintf("%s├─ Base: %s\n", indent, r.formatBaseExprRecursive(expr.Base, indent+"│  "))

	for i, access := range expr.Access {
		prefix := "├─ "
		childIndent := indent + "│  "
		if i == len(expr.Access)-1 {
			prefix = "└─ "
			childIndent = indent + "   "
		}

		if access.MethodCall != nil {
			result += fmt.Sprintf("%s%sMethod: '%s'\n", indent, prefix, access.MethodCall.Method)
			if len(access.MethodCall.Args) > 0 {
				result += fmt.Sprintf("%s   └─ Args:\n", indent)
				for j, arg := range access.MethodCall.Args {
					argPrefix := "├─ "
					if j == len(access.MethodCall.Args)-1 {
						argPrefix = "└─ "
					}
					result += fmt.Sprintf("%s      %s%s", indent, argPrefix, r.formatExpressionRecursive(arg, childIndent+"      "))
				}
			}
		}
	}

	return result
}

// formatBaseExprRecursive formats base expressions recursively
func (r *REPL) formatBaseExprRecursive(expr *parser.BaseExpr, indent string) string {
	switch {
	case expr.Literal != nil:
		return r.formatLiteralRecursive(expr.Literal, indent)
	case expr.Identifier != nil:
		return fmt.Sprintf("Identifier: %s", *expr.Identifier)
	case expr.StructConstructor != nil:
		s := expr.StructConstructor
		result := fmt.Sprintf("StructConstructor: %s\n", s.Name)
		for i, field := range s.Fields {
			prefix := "├─ "
			if i == len(s.Fields)-1 {
				prefix = "└─ "
			}
			result += fmt.Sprintf("%s%s%s: %s\n", indent, prefix, field.Key, r.formatExpressionRecursive(field.Value, indent+"   "))
		}
		return result
	case expr.ObjectLit != nil:
		o := expr.ObjectLit
		result := "ObjectLiteral:\n"
		for i, field := range o.Fields {
			prefix := "├─ "
			if i == len(o.Fields)-1 {
				prefix = "└─ "
			}
			result += fmt.Sprintf("%s%s%s: %s\n", indent, prefix, field.Key, r.formatExpressionRecursive(field.Value, indent+"   "))
		}
		return result
	case expr.SendExpr != nil:
		s := expr.SendExpr
		result := fmt.Sprintf("SendExpr: '%s'.%s\n", s.Target, s.Method)
		if s.Args != nil {
			result += fmt.Sprintf("%s└─ Args: %s", indent, r.formatBaseExprRecursive(&parser.BaseExpr{ObjectLit: s.Args}, indent+"   "))
		}
		return result
	case expr.Lambda != nil:
		l := expr.Lambda
		result := "Lambda:\n"
		if len(l.Parameters) > 0 {
			result += fmt.Sprintf("%s├─ Parameters:\n", indent)
			for i, param := range l.Parameters {
				prefix := "├─ "
				if i == len(l.Parameters)-1 {
					prefix = "└─ "
				}
				result += fmt.Sprintf("%s│  %s%s: %s\n", indent, prefix, param.Name, r.formatTypeRefRecursive(param.Type, indent+"│  "))
			}
		}
		if l.ReturnType != nil {
			result += fmt.Sprintf("%s├─ ReturnType: %s\n", indent, r.formatTypeRefRecursive(l.ReturnType, indent+"│  "))
		}
		if l.Body != nil {
			result += fmt.Sprintf("%s└─ Body:\n", indent)
			result += r.formatBlockRecursive(l.Body, indent+"   ")
		}
		return result
	case expr.FuncCall != nil:
		f := expr.FuncCall
		result := fmt.Sprintf("FuncCall: %s\n", f.Name)
		if len(f.Args) > 0 {
			result += fmt.Sprintf("%s└─ Args:\n", indent)
			for i, arg := range f.Args {
				prefix := "├─ "
				if i == len(f.Args)-1 {
					prefix = "└─ "
				}
				result += fmt.Sprintf("%s   %s%s\n", indent, prefix, r.formatExpressionRecursive(arg, indent+"      "))
			}
		}
		return result
	case expr.Block != nil:
		result := "Block:\n"
		result += r.formatBlockRecursive(expr.Block, indent)
		return result
	case expr.Grouped != nil:
		return fmt.Sprintf("Grouped: (%s)", r.formatExpressionRecursive(expr.Grouped, indent))
	default:
		return "UnknownBaseExpr"
	}
}

// handleCommand processes special REPL commands
func (r *REPL) handleCommand(cmd string) {
	switch strings.ToLower(strings.TrimSpace(cmd)) {
	case ":help", ":h":
		r.showHelp()
	case ":quit", ":q", ":exit":
		fmt.Fprintln(r.output, "Goodbye!")
		os.Exit(0)
	case ":clear", ":cls":
		r.clearScreen()
	case ":examples", ":ex":
		r.showExamples()
	case ":ast":
		r.showASTHelp()
	case ":test":
		r.runQuickTests()
	case ":version", ":v":
		r.showVersion()
	case ":mode":
		r.toggleMode()
	case ":exec":
		r.showAST = false
		fmt.Fprintln(r.output, "Switched to execution mode")
	case ":astmode":
		r.showAST = true
		fmt.Fprintln(r.output, "Switched to AST viewing mode")
	default:
		fmt.Fprintf(r.output, "Unknown command: %s\n", cmd)
		fmt.Fprintln(r.output, "Type :help for available commands")
	}
}

// toggleMode switches between execution and AST modes
func (r *REPL) toggleMode() {
	r.showAST = !r.showAST
	if r.showAST {
		fmt.Fprintln(r.output, "Switched to AST viewing mode")
	} else {
		fmt.Fprintln(r.output, "Switched to execution mode")
	}
}

// showHelp displays available REPL commands
func (r *REPL) showHelp() {
	fmt.Fprintln(r.output, "=== Relay REPL Commands ===")
	fmt.Fprintln(r.output, ":help, :h     - Show this help")
	fmt.Fprintln(r.output, ":quit, :q     - Quit the REPL")
	fmt.Fprintln(r.output, ":exit         - Quit the REPL")
	fmt.Fprintln(r.output, ":clear, :c    - Clear the screen")
	fmt.Fprintln(r.output, ":examples, :ex - Show example code")
	fmt.Fprintln(r.output, ":ast          - Show AST parsing tips")
	fmt.Fprintln(r.output, ":test         - Run quick syntax tests")
	fmt.Fprintln(r.output, ":version, :v  - Show version info")
	fmt.Fprintln(r.output, ":mode         - Toggle between execution and AST modes")
	fmt.Fprintln(r.output, ":exec         - Switch to execution mode")
	fmt.Fprintln(r.output, ":astmode      - Switch to AST viewing mode")
	fmt.Fprintln(r.output, "===========================")
}

// clearScreen clears the terminal screen
func (r *REPL) clearScreen() {
	fmt.Fprint(r.output, "\033[2J\033[H")
}

// showExamples displays example Relay code
func (r *REPL) showExamples() {
	fmt.Fprintln(r.output, "=== Example Relay Code ===")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "1. Simple struct:")
	fmt.Fprintln(r.output, "   struct User {")
	fmt.Fprintln(r.output, "       name: string,")
	fmt.Fprintln(r.output, "       email: string")
	fmt.Fprintln(r.output, "   }")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "2. Function definition:")
	fmt.Fprintln(r.output, "   fn greet(name: string) -> string {")
	fmt.Fprintln(r.output, "       \"Hello, \" + name")
	fmt.Fprintln(r.output, "   }")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "3. Variable assignment:")
	fmt.Fprintln(r.output, "   set message = \"Hello, world!\"")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "4. Protocol definition:")
	fmt.Fprintln(r.output, "   protocol UserService {")
	fmt.Fprintln(r.output, "       get_user(id: string) -> User")
	fmt.Fprintln(r.output, "   }")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "5. Lambda expression:")
	fmt.Fprintln(r.output, "   set doubled = numbers.map(fn (x) { x * 2 })")
	fmt.Fprintln(r.output, "==========================")
}

// showASTHelp displays information about AST inspection
func (r *REPL) showASTHelp() {
	fmt.Fprintln(r.output, "=== AST Inspection Help ===")
	fmt.Fprintln(r.output, "The REPL shows parsed AST structures for all valid Relay code.")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "Structure types you'll see:")
	fmt.Fprintln(r.output, "• Struct 'Name' with N fields")
	fmt.Fprintln(r.output, "• Protocol 'Name' with N methods")
	fmt.Fprintln(r.output, "• Server 'Name'")
	fmt.Fprintln(r.output, "• Function 'name' with N parameters")
	fmt.Fprintln(r.output, "• Set variable 'name'")
	fmt.Fprintln(r.output, "• Logical/If/For/Try expressions")
	fmt.Fprintln(r.output, "")
	fmt.Fprintln(r.output, "This helps you understand how the parser")
	fmt.Fprintln(r.output, "interprets your Relay code structure.")
	fmt.Fprintln(r.output, "===========================")
}

// runQuickTests runs some basic syntax validation tests
func (r *REPL) runQuickTests() {
	fmt.Fprintln(r.output, "=== Running Quick Syntax Tests ===")

	tests := []struct {
		name string
		code string
	}{
		{"Simple struct", `struct Test { name: string }`},
		{"Protocol", `protocol TestService { test() -> string }`},
		{"Function", `fn test() -> string { "hello" }`},
		{"Variable", `set x = 42`},
		{"Lambda", `set f = fn(x) { x + 1 }`},
	}

	passed := 0
	for _, test := range tests {
		_, err := parser.Parse("test", strings.NewReader(test.code))
		if err != nil {
			fmt.Fprintf(r.output, "❌ %s: %v\n", test.name, err)
		} else {
			fmt.Fprintf(r.output, "✅ %s: OK\n", test.name)
			passed++
		}
	}

	fmt.Fprintf(r.output, "\nPassed: %d/%d tests\n", passed, len(tests))
	fmt.Fprintln(r.output, "===============================")
}

// showVersion displays version information
func (r *REPL) showVersion() {
	fmt.Fprintln(r.output, "=== Relay Language Info ===")
	fmt.Fprintln(r.output, "Version: 0.3.0-dev")
	fmt.Fprintln(r.output, "Parser: Participle v2.1.4")
	fmt.Fprintln(r.output, "Status: Development")
	fmt.Fprintln(r.output, "===========================")
}

// welcome returns the welcome message
func welcome() string {
	return `
╔═══════════════════════════════════════╗
║        Welcome to Relay REPL!        ║
║   Federated Web Programming Language  ║
║                                       ║
║  🚀 Now with Runtime Execution!      ║
╚═══════════════════════════════════════╝

Mode: EXECUTION (evaluates code)
Type :help for available commands
Type :mode to toggle between execution and AST viewing
Type :examples to see sample code
Press Ctrl+C or :quit to exit

`
}
