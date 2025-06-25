# Go LLVM Introduction and Cheatsheet

## Overview

This guide provides a practical introduction to using LLVM from Go, with examples specifically relevant to the Relay language project. LLVM (Low Level Virtual Machine) is a compiler infrastructure that allows you to generate optimized machine code at runtime.

## LLVM Go Libraries

### Option 1: llir/llvm (Pure Go - Recommended for Relay)
```bash
go get github.com/llir/llvm@latest
```

**Pros:**
- Pure Go implementation
- No CGO dependencies
- Easy to cross-compile
- Good for IR generation and manipulation

**Cons:**
- Cannot execute code directly (IR generation only)
- Need external tools for compilation to machine code

### Option 2: go-llvm (CGO Bindings)
```bash
# Requires LLVM development libraries installed
# macOS: brew install llvm
# Ubuntu: apt-get install llvm-dev
go get github.com/llir/llvm@latest
```

**Pros:**
- Full LLVM functionality including JIT compilation
- Can execute generated code directly

**Cons:**
- Requires CGO and LLVM C++ libraries
- More complex build process

## Basic Concepts

### LLVM IR (Intermediate Representation)
LLVM IR is a low-level programming language similar to assembly, but target-independent.

**Key Components:**
- **Module**: Top-level container (like a source file)
- **Function**: Contains basic blocks and instructions
- **Basic Block**: Sequence of instructions with single entry/exit
- **Instruction**: Individual operations (add, load, store, etc.)
- **Value**: Represents data (constants, variables, instruction results)

## Getting Started with llir/llvm

### 1. Basic Setup

```go
package main

import (
    "fmt"
    "github.com/llir/llvm/ir"
    "github.com/llir/llvm/ir/constant"
    "github.com/llir/llvm/ir/types"
)

func main() {
    // Create a new LLVM module
    m := ir.NewModule()
    
    // Print the empty module
    fmt.Println(m)
}
```

### 2. Creating a Simple Function

```go
package main

import (
    "fmt"
    "github.com/llir/llvm/ir"
    "github.com/llir/llvm/ir/constant"
    "github.com/llir/llvm/ir/types"
)

func createAddFunction() *ir.Module {
    // Create module
    m := ir.NewModule()
    
    // Create function signature: i32 add(i32 %a, i32 %b)
    funcType := types.NewFunc(types.I32, types.I32, types.I32)
    addFunc := m.NewFunc("add", funcType)
    
    // Name the parameters
    a := addFunc.Params[0]
    a.SetName("a")
    b := addFunc.Params[1]
    b.SetName("b")
    
    // Create entry basic block
    entry := addFunc.NewBlock("entry")
    
    // Add instruction: %result = add i32 %a, %b
    result := entry.NewAdd(a, b)
    result.SetName("result")
    
    // Return the result
    entry.NewRet(result)
    
    return m
}

func main() {
    module := createAddFunction()
    fmt.Println(module)
}
```

**Output LLVM IR:**
```llvm
define i32 @add(i32 %a, i32 %b) {
entry:
  %result = add i32 %a, %b
  ret i32 %result
}
```

### 3. Working with Constants

```go
func createConstantExample() *ir.Module {
    m := ir.NewModule()
    
    // Create function: i32 getAnswer()
    funcType := types.NewFunc(types.I32)
    getAnswer := m.NewFunc("getAnswer", funcType)
    
    entry := getAnswer.NewBlock("entry")
    
    // Create constant
    answer := constant.NewInt(types.I32, 42)
    
    // Return constant
    entry.NewRet(answer)
    
    return m
}
```

### 4. Control Flow (If/Else)

```go
func createIfElseExample() *ir.Module {
    m := ir.NewModule()
    
    // Function: i32 abs(i32 %x)
    funcType := types.NewFunc(types.I32, types.I32)
    absFunc := m.NewFunc("abs", funcType)
    
    x := absFunc.Params[0]
    x.SetName("x")
    
    // Basic blocks
    entry := absFunc.NewBlock("entry")
    ifNeg := absFunc.NewBlock("if_negative")
    ifPos := absFunc.NewBlock("if_positive")
    exit := absFunc.NewBlock("exit")
    
    // Entry: compare x with 0
    zero := constant.NewInt(types.I32, 0)
    cond := entry.NewICmp(enum.IPredSLT, x, zero) // x < 0
    entry.NewCondBr(cond, ifNeg, ifPos)
    
    // If negative: negate x
    negX := ifNeg.NewSub(zero, x)
    ifNeg.NewBr(exit)
    
    // If positive: use x as-is
    ifPos.NewBr(exit)
    
    // Exit: phi node to select result
    result := exit.NewPhi(ir.NewIncoming(negX, ifNeg), ir.NewIncoming(x, ifPos))
    exit.NewRet(result)
    
    return m
}
```

### 5. Loops

```go
func createLoopExample() *ir.Module {
    m := ir.NewModule()
    
    // Function: i32 sum_to_n(i32 %n)
    funcType := types.NewFunc(types.I32, types.I32)
    sumFunc := m.NewFunc("sum_to_n", funcType)
    
    n := sumFunc.Params[0]
    n.SetName("n")
    
    // Basic blocks
    entry := sumFunc.NewBlock("entry")
    loop := sumFunc.NewBlock("loop")
    exit := sumFunc.NewBlock("exit")
    
    // Constants
    zero := constant.NewInt(types.I32, 0)
    one := constant.NewInt(types.I32, 1)
    
    // Entry: initialize and jump to loop
    entry.NewBr(loop)
    
    // Loop header: phi nodes for i and sum
    i := loop.NewPhi(ir.NewIncoming(zero, entry))
    sum := loop.NewPhi(ir.NewIncoming(zero, entry))
    
    // Loop body: sum += i, i++
    newSum := loop.NewAdd(sum, i)
    newI := loop.NewAdd(i, one)
    
    // Update phi nodes
    i.Incs = append(i.Incs, ir.NewIncoming(newI, loop))
    sum.Incs = append(sum.Incs, ir.NewIncoming(newSum, loop))
    
    // Loop condition: i < n
    cond := loop.NewICmp(enum.IPredSLT, newI, n)
    loop.NewCondBr(cond, loop, exit)
    
    // Exit
    exit.NewRet(newSum)
    
    return m
}
```

## Common LLVM Types in Go

```go
// Integer types
types.I1    // bool
types.I8    // byte
types.I32   // int32
types.I64   // int64

// Floating point types
types.Float   // float32
types.Double  // float64

// Pointer types
types.NewPointer(types.I32)  // *i32

// Array types
types.NewArray(10, types.I32)  // [10 x i32]

// Function types
types.NewFunc(types.I32, types.I32, types.I32)  // i32(i32, i32)

// Struct types
types.NewStruct(types.I32, types.Double)  // {i32, double}
```

## Common Instructions Cheatsheet

### Arithmetic Operations
```go
// Integer arithmetic
entry.NewAdd(a, b)    // a + b
entry.NewSub(a, b)    // a - b
entry.NewMul(a, b)    // a * b
entry.NewSDiv(a, b)   // a / b (signed)
entry.NewSRem(a, b)   // a % b (signed)

// Floating point arithmetic
entry.NewFAdd(a, b)   // a + b (float)
entry.NewFSub(a, b)   // a - b (float)
entry.NewFMul(a, b)   // a * b (float)
entry.NewFDiv(a, b)   // a / b (float)
```

### Memory Operations
```go
// Allocate stack memory
ptr := entry.NewAlloca(types.I32)

// Store value to memory
entry.NewStore(value, ptr)

// Load value from memory
loaded := entry.NewLoad(types.I32, ptr)

// Get element pointer (for arrays/structs)
gep := entry.NewGetElementPtr(types.I32, arrayPtr, index)
```

### Comparison Operations
```go
// Integer comparisons
entry.NewICmp(enum.IPredEQ, a, b)   // a == b
entry.NewICmp(enum.IPredNE, a, b)   // a != b
entry.NewICmp(enum.IPredSLT, a, b)  // a < b (signed)
entry.NewICmp(enum.IPredSLE, a, b)  // a <= b (signed)
entry.NewICmp(enum.IPredSGT, a, b)  // a > b (signed)
entry.NewICmp(enum.IPredSGE, a, b)  // a >= b (signed)

// Float comparisons
entry.NewFCmp(enum.FPredOEQ, a, b)  // a == b (float)
entry.NewFCmp(enum.FPredOLT, a, b)  // a < b (float)
```

### Control Flow
```go
// Unconditional branch
entry.NewBr(targetBlock)

// Conditional branch
entry.NewCondBr(condition, trueBlock, falseBlock)

// Return
entry.NewRet(value)
entry.NewRet(nil)  // void return

// Function call
result := entry.NewCall(function, arg1, arg2)
```

## Relay-Specific Examples

### 1. Compiling Relay Number Literals

```go
func compileRelayNumber(value float64) *ir.Module {
    m := ir.NewModule()
    
    // Function: double getNumber()
    funcType := types.NewFunc(types.Double)
    getNum := m.NewFunc("getNumber", funcType)
    
    entry := getNum.NewBlock("entry")
    
    // Create constant from Relay number
    num := constant.NewFloat(types.Double, value)
    entry.NewRet(num)
    
    return m
}

// Usage for Relay: let x = 42.5
module := compileRelayNumber(42.5)
```

### 2. Compiling Relay Binary Expressions

```go
func compileRelayBinaryOp(left, right float64, op string) *ir.Module {
    m := ir.NewModule()
    
    funcType := types.NewFunc(types.Double)
    calc := m.NewFunc("calculate", funcType)
    
    entry := calc.NewBlock("entry")
    
    leftVal := constant.NewFloat(types.Double, left)
    rightVal := constant.NewFloat(types.Double, right)
    
    var result value.Value
    switch op {
    case "+":
        result = entry.NewFAdd(leftVal, rightVal)
    case "-":
        result = entry.NewFSub(leftVal, rightVal)
    case "*":
        result = entry.NewFMul(leftVal, rightVal)
    case "/":
        result = entry.NewFDiv(leftVal, rightVal)
    default:
        panic("unsupported operator")
    }
    
    entry.NewRet(result)
    return m
}

// Usage for Relay: 10.5 + 20.3
module := compileRelayBinaryOp(10.5, 20.3, "+")
```

### 3. Compiling Relay Functions

```go
func compileRelayFunction(name string, paramNames []string, body string) *ir.Module {
    m := ir.NewModule()
    
    // Create parameter types (assume all double for simplicity)
    paramTypes := make([]types.Type, len(paramNames))
    for i := range paramTypes {
        paramTypes[i] = types.Double
    }
    
    // Function type
    funcType := types.NewFunc(types.Double, paramTypes...)
    fn := m.NewFunc(name, funcType)
    
    // Name parameters
    for i, paramName := range paramNames {
        fn.Params[i].SetName(paramName)
    }
    
    entry := fn.NewBlock("entry")
    
    // For this example, just return first parameter
    // In real implementation, you'd compile the body AST
    if len(fn.Params) > 0 {
        entry.NewRet(fn.Params[0])
    } else {
        zero := constant.NewFloat(types.Double, 0.0)
        entry.NewRet(zero)
    }
    
    return m
}

// Usage for Relay: fn double(x) { x * 2 }
module := compileRelayFunction("double", []string{"x"}, "x * 2")
```

### 4. Relay Actor Message Handling

```go
func compileRelayReceiveFunction() *ir.Module {
    m := ir.NewModule()
    
    // Simulate: receive fn process(data) { data + 1 }
    // Function signature: double process(double data)
    funcType := types.NewFunc(types.Double, types.Double)
    process := m.NewFunc("process", funcType)
    
    data := process.Params[0]
    data.SetName("data")
    
    entry := process.NewBlock("entry")
    
    // Add 1 to data
    one := constant.NewFloat(types.Double, 1.0)
    result := entry.NewFAdd(data, one)
    
    entry.NewRet(result)
    
    return m
}
```

## Debugging and Validation

### 1. Print Generated IR

```go
func printIR(m *ir.Module) {
    fmt.Println("Generated LLVM IR:")
    fmt.Println(m.String())
}
```

### 2. Validate Module

```go
import "github.com/llir/llvm/ir/irutil"

func validateModule(m *ir.Module) error {
    // Basic validation
    for _, fn := range m.Funcs {
        if len(fn.Blocks) == 0 {
            return fmt.Errorf("function %s has no basic blocks", fn.Name())
        }
        
        for _, block := range fn.Blocks {
            if len(block.Insts) == 0 {
                return fmt.Errorf("basic block %s is empty", block.Name())
            }
            
            // Check if block has terminator
            lastInst := block.Insts[len(block.Insts)-1]
            if !irutil.IsTerminator(lastInst) {
                return fmt.Errorf("basic block %s missing terminator", block.Name())
            }
        }
    }
    return nil
}
```

## Complete Example: Relay Expression Compiler

```go
package main

import (
    "fmt"
    "github.com/llir/llvm/ir"
    "github.com/llir/llvm/ir/constant"
    "github.com/llir/llvm/ir/types"
    "github.com/llir/llvm/ir/enum"
)

// Simple AST nodes for Relay expressions
type RelayExpr interface {
    String() string
}

type RelayNumber struct {
    Value float64
}

func (n *RelayNumber) String() string {
    return fmt.Sprintf("%.2f", n.Value)
}

type RelayBinary struct {
    Left  RelayExpr
    Op    string
    Right RelayExpr
}

func (b *RelayBinary) String() string {
    return fmt.Sprintf("(%s %s %s)", b.Left, b.Op, b.Right)
}

type RelayIf struct {
    Condition RelayExpr
    ThenExpr  RelayExpr
    ElseExpr  RelayExpr
}

func (i *RelayIf) String() string {
    return fmt.Sprintf("if %s then %s else %s", i.Condition, i.ThenExpr, i.ElseExpr)
}

// Compiler for Relay expressions
type RelayCompiler struct {
    module   *ir.Module
    function *ir.Func
    builder  *ir.Block
}

func NewRelayCompiler() *RelayCompiler {
    m := ir.NewModule()
    
    // Create main function
    funcType := types.NewFunc(types.Double)
    mainFunc := m.NewFunc("main", funcType)
    
    entry := mainFunc.NewBlock("entry")
    
    return &RelayCompiler{
        module:   m,
        function: mainFunc,
        builder:  entry,
    }
}

func (c *RelayCompiler) Compile(expr RelayExpr) (*ir.Module, error) {
    result, err := c.compileExpr(expr)
    if err != nil {
        return nil, err
    }
    
    c.builder.NewRet(result)
    return c.module, nil
}

func (c *RelayCompiler) compileExpr(expr RelayExpr) (value.Value, error) {
    switch e := expr.(type) {
    case *RelayNumber:
        return constant.NewFloat(types.Double, e.Value), nil
        
    case *RelayBinary:
        left, err := c.compileExpr(e.Left)
        if err != nil {
            return nil, err
        }
        
        right, err := c.compileExpr(e.Right)
        if err != nil {
            return nil, err
        }
        
        switch e.Op {
        case "+":
            return c.builder.NewFAdd(left, right), nil
        case "-":
            return c.builder.NewFSub(left, right), nil
        case "*":
            return c.builder.NewFMul(left, right), nil
        case "/":
            return c.builder.NewFDiv(left, right), nil
        case "<":
            return c.builder.NewFCmp(enum.FPredOLT, left, right), nil
        case ">":
            return c.builder.NewFCmp(enum.FPredOGT, left, right), nil
        case "==":
            return c.builder.NewFCmp(enum.FPredOEQ, left, right), nil
        default:
            return nil, fmt.Errorf("unsupported operator: %s", e.Op)
        }
        
    case *RelayIf:
        // Compile condition
        cond, err := c.compileExpr(e.Condition)
        if err != nil {
            return nil, err
        }
        
        // Create basic blocks
        thenBlock := c.function.NewBlock("then")
        elseBlock := c.function.NewBlock("else")
        mergeBlock := c.function.NewBlock("merge")
        
        // Conditional branch
        c.builder.NewCondBr(cond, thenBlock, elseBlock)
        
        // Compile then branch
        c.builder = thenBlock
        thenVal, err := c.compileExpr(e.ThenExpr)
        if err != nil {
            return nil, err
        }
        c.builder.NewBr(mergeBlock)
        thenBlock = c.builder // Update in case new blocks were created
        
        // Compile else branch
        c.builder = elseBlock
        elseVal, err := c.compileExpr(e.ElseExpr)
        if err != nil {
            return nil, err
        }
        c.builder.NewBr(mergeBlock)
        elseBlock = c.builder // Update in case new blocks were created
        
        // Merge block with phi
        c.builder = mergeBlock
        phi := c.builder.NewPhi(
            ir.NewIncoming(thenVal, thenBlock),
            ir.NewIncoming(elseVal, elseBlock),
        )
        
        return phi, nil
        
    default:
        return nil, fmt.Errorf("unsupported expression type: %T", expr)
    }
}

func main() {
    // Example: if (10.0 > 5.0) then (20.0 + 30.0) else (40.0 - 10.0)
    expr := &RelayIf{
        Condition: &RelayBinary{
            Left:  &RelayNumber{Value: 10.0},
            Op:    ">",
            Right: &RelayNumber{Value: 5.0},
        },
        ThenExpr: &RelayBinary{
            Left:  &RelayNumber{Value: 20.0},
            Op:    "+",
            Right: &RelayNumber{Value: 30.0},
        },
        ElseExpr: &RelayBinary{
            Left:  &RelayNumber{Value: 40.0},
            Op:    "-",
            Right: &RelayNumber{Value: 10.0},
        },
    }
    
    compiler := NewRelayCompiler()
    module, err := compiler.Compile(expr)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Relay Expression: %s\n\n", expr)
    fmt.Println("Generated LLVM IR:")
    fmt.Println(module)
}
```

## Next Steps for Relay Integration

1. **Start with llir/llvm**: Use the pure Go library for IR generation
2. **Create type mapping**: Map Relay types to LLVM types
3. **Implement AST visitors**: Convert Relay AST nodes to LLVM IR
4. **Add optimization passes**: Use LLVM's optimization capabilities
5. **Consider JIT compilation**: Later add go-llvm for runtime execution

## Useful Resources

- [llir/llvm Documentation](https://pkg.go.dev/github.com/llir/llvm)
- [LLVM Language Reference](https://llvm.org/docs/LangRef.html)
- [LLVM IR Tutorial](https://llvm.org/docs/tutorial/)
- [Go LLVM Examples](https://github.com/llir/llvm/tree/master/example)

This cheatsheet should give you a solid foundation for implementing LLVM compilation in the Relay language project! 