package parser

import (
	"strings"
	"testing"
)

func TestParser_SetStatements(t *testing.T) {
	code := `receive test {} -> string {
		set x = "hello"
		set y = 42
		set z = true
		return x
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	if receive.Name != "test" {
		t.Errorf("Expected method name 'test', got '%s'", receive.Name)
	}

	if len(receive.Body.Statements) != 4 {
		t.Fatalf("Expected 4 body statements, got %d", len(receive.Body.Statements))
	}

	// Test string set statement
	setStmt1 := receive.Body.Statements[0].SetStatement
	if setStmt1 == nil {
		t.Fatal("Expected first statement to be set statement")
	}
	if setStmt1.Variable != "x" {
		t.Errorf("Expected variable 'x', got '%s'", setStmt1.Variable)
	}

	// Test number set statement
	setStmt2 := receive.Body.Statements[1].SetStatement
	if setStmt2 == nil {
		t.Fatal("Expected second statement to be set statement")
	}
	if setStmt2.Variable != "y" {
		t.Errorf("Expected variable 'y', got '%s'", setStmt2.Variable)
	}

	// Test boolean set statement
	setStmt3 := receive.Body.Statements[2].SetStatement
	if setStmt3 == nil {
		t.Fatal("Expected third statement to be set statement")
	}
	if setStmt3.Variable != "z" {
		t.Errorf("Expected variable 'z', got '%s'", setStmt3.Variable)
	}

	// Test return statement
	returnStmt := receive.Body.Statements[3].ReturnStatement
	if returnStmt == nil {
		t.Fatal("Expected fourth statement to be return statement")
	}
}

func TestParser_BinaryOperations(t *testing.T) {
	code := `receive test {} -> number {
		set sum = a + b
		set diff = x - y
		set product = m * n
		set quotient = p / q
		return sum
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	// Test addition
	setStmt1 := receive.Body.Statements[0].SetStatement
	if setStmt1 == nil {
		t.Fatal("Expected set statement")
	}
	addExpr := setStmt1.Value.Logical.Left.Left.Left
	if addExpr == nil || len(addExpr.Right) == 0 {
		t.Fatal("Expected additive expression with operation")
	}
	if addExpr.Right[0].Op != "+" {
		t.Errorf("Expected '+', got '%s'", addExpr.Right[0].Op)
	}

	// Test subtraction
	setStmt2 := receive.Body.Statements[1].SetStatement
	if setStmt2 == nil {
		t.Fatal("Expected set statement")
	}
	subExpr := setStmt2.Value.Logical.Left.Left.Left
	if subExpr == nil || len(subExpr.Right) == 0 {
		t.Fatal("Expected additive expression with operation")
	}
	if subExpr.Right[0].Op != "-" {
		t.Errorf("Expected '-', got '%s'", subExpr.Right[0].Op)
	}

	// Test multiplication
	setStmt3 := receive.Body.Statements[2].SetStatement
	if setStmt3 == nil {
		t.Fatal("Expected set statement")
	}
	mulExpr := setStmt3.Value.Logical.Left.Left.Left.Left
	if mulExpr == nil || len(mulExpr.Right) == 0 {
		t.Fatal("Expected multiplicative expression with operation")
	}
	if mulExpr.Right[0].Op != "*" {
		t.Errorf("Expected '*', got '%s'", mulExpr.Right[0].Op)
	}

	// Test division
	setStmt4 := receive.Body.Statements[3].SetStatement
	if setStmt4 == nil {
		t.Fatal("Expected set statement")
	}
	divExpr := setStmt4.Value.Logical.Left.Left.Left.Left
	if divExpr == nil || len(divExpr.Right) == 0 {
		t.Fatal("Expected multiplicative expression with operation")
	}
	if divExpr.Right[0].Op != "/" {
		t.Errorf("Expected '/', got '%s'", divExpr.Right[0].Op)
	}
}

func TestParser_ComparisonOperations(t *testing.T) {
	code := `receive test {} -> bool {
		set eq = a == b
		set ne = x != y
		set lt = m < n
		set gt = p > q
		return eq
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	tests := []struct {
		index    int
		expected string
	}{
		{0, "=="}, // equality
		{1, "!="}, // inequality
		{2, "<"},  // less than
		{3, ">"},  // greater than
	}

	for _, test := range tests {
		setStmt := receive.Body.Statements[test.index].SetStatement
		if setStmt == nil {
			t.Fatalf("Expected set statement at index %d", test.index)
		}

		var op string
		if test.expected == "==" || test.expected == "!=" {
			// Equality operations
			eqExpr := setStmt.Value.Logical.Left
			if eqExpr == nil || len(eqExpr.Right) == 0 {
				t.Fatalf("Expected equality expression with operation at index %d", test.index)
			}
			op = eqExpr.Right[0].Op
		} else {
			// Relational operations
			relExpr := setStmt.Value.Logical.Left.Left
			if relExpr == nil || len(relExpr.Right) == 0 {
				t.Fatalf("Expected relational expression with operation at index %d", test.index)
			}
			op = relExpr.Right[0].Op
		}

		if op != test.expected {
			t.Errorf("Statement %d: Expected '%s', got '%s'", test.index, test.expected, op)
		}
	}
}

func TestParser_LogicalOperations(t *testing.T) {
	code := `receive test {} -> bool {
		set and_expr = a && b
		set or_expr = x || y
		return and_expr
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	// Test AND operation
	setStmt1 := receive.Body.Statements[0].SetStatement
	if setStmt1 == nil {
		t.Fatal("Expected set statement")
	}
	logicalExpr := setStmt1.Value.Logical
	if logicalExpr == nil || len(logicalExpr.Right) == 0 {
		t.Fatal("Expected logical expression with operation")
	}
	if logicalExpr.Right[0].Op != "&&" {
		t.Errorf("Expected '&&', got '%s'", logicalExpr.Right[0].Op)
	}

	// Test OR operation
	setStmt2 := receive.Body.Statements[1].SetStatement
	if setStmt2 == nil {
		t.Fatal("Expected set statement")
	}
	logicalExpr2 := setStmt2.Value.Logical
	if logicalExpr2 == nil || len(logicalExpr2.Right) == 0 {
		t.Fatal("Expected logical expression with operation")
	}
	if logicalExpr2.Right[0].Op != "||" {
		t.Errorf("Expected '||', got '%s'", logicalExpr2.Right[0].Op)
	}
}

func TestParser_FieldAccess(t *testing.T) {
	code := `receive test {} -> string {
		set name = user.name
		return name
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	// Test simple field access
	setStmt1 := receive.Body.Statements[0].SetStatement
	if setStmt1 == nil {
		t.Fatal("Expected set statement")
	}

	// Navigate to the primary expression through UnaryExpr
	primaryExpr := setStmt1.Value.Logical.Left.Left.Left.Left.Left.Primary
	if primaryExpr == nil {
		t.Fatal("Expected primary expression")
	}

	if primaryExpr.Base.Identifier == nil {
		t.Fatal("Expected base identifier")
	}
	if *primaryExpr.Base.Identifier != "user" {
		t.Errorf("Expected 'user', got '%s'", *primaryExpr.Base.Identifier)
	}
	if len(primaryExpr.Access) == 0 || primaryExpr.Access[0].FieldAccess == nil {
		t.Fatal("Expected field access")
	}
	if *primaryExpr.Access[0].FieldAccess != "name" {
		t.Errorf("Expected 'name', got '%s'", *primaryExpr.Access[0].FieldAccess)
	}
}

func TestParser_ObjectLiterals(t *testing.T) {
	code := `receive test {} -> object {
		set obj = {name: "John", age: 30}
		set empty = {}
		return obj
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	// Test object literal with fields
	setStmt1 := receive.Body.Statements[0].SetStatement
	if setStmt1 == nil {
		t.Fatal("Expected set statement")
	}

	// Navigate to object literal through UnaryExpr
	objectLit := setStmt1.Value.Logical.Left.Left.Left.Left.Left.Primary.Base.ObjectLit
	if objectLit == nil {
		t.Fatal("Expected object literal")
	}
	if len(objectLit.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(objectLit.Fields))
	}
	if objectLit.Fields[0].Key != "name" {
		t.Errorf("Expected field 'name', got '%s'", objectLit.Fields[0].Key)
	}
	if objectLit.Fields[1].Key != "age" {
		t.Errorf("Expected field 'age', got '%s'", objectLit.Fields[1].Key)
	}

	// Test empty object literal
	setStmt2 := receive.Body.Statements[1].SetStatement
	if setStmt2 == nil {
		t.Fatal("Expected set statement")
	}
	objectLit2 := setStmt2.Value.Logical.Left.Left.Left.Left.Left.Primary.Base.ObjectLit
	if objectLit2 == nil {
		t.Fatal("Expected object literal")
	}
	if len(objectLit2.Fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(objectLit2.Fields))
	}
}

func TestParser_IfStatements(t *testing.T) {
	code := `receive test {} -> string {
		if condition {
			return "true"
		}
		if x == y {
			set result = "equal"
		} else {
			set result = "not equal"
		}
		return result
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	// Test simple if statement
	ifStmt1 := receive.Body.Statements[0].IfStatement
	if ifStmt1 == nil {
		t.Fatal("Expected if statement")
	}
	if ifStmt1.Condition == nil {
		t.Fatal("Expected if condition")
	}
	if ifStmt1.ThenBlock == nil {
		t.Fatal("Expected then block")
	}
	if ifStmt1.ElseBlock != nil {
		t.Error("Expected no else block")
	}
	if len(ifStmt1.ThenBlock.Statements) != 1 {
		t.Errorf("Expected 1 statement in then block, got %d", len(ifStmt1.ThenBlock.Statements))
	}

	// Test if-else statement
	ifStmt2 := receive.Body.Statements[1].IfStatement
	if ifStmt2 == nil {
		t.Fatal("Expected if statement")
	}
	if ifStmt2.Condition == nil {
		t.Fatal("Expected if condition")
	}
	if ifStmt2.ThenBlock == nil {
		t.Fatal("Expected then block")
	}
	if ifStmt2.ElseBlock == nil {
		t.Fatal("Expected else block")
	}
	if len(ifStmt2.ThenBlock.Statements) != 1 {
		t.Errorf("Expected 1 statement in then block, got %d", len(ifStmt2.ThenBlock.Statements))
	}
	if len(ifStmt2.ElseBlock.Statements) != 1 {
		t.Errorf("Expected 1 statement in else block, got %d", len(ifStmt2.ElseBlock.Statements))
	}
}

func TestParser_MixedExpressions(t *testing.T) {
	code := `receive process {} -> object {
		set user = {
			name: "John",
			age: 30
		}
		
		if user.age > 17 {
			set status = "adult"
		}
		
		return user
	}`

	program, err := Parse("test", strings.NewReader(code))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	receive := program.Statements[0].ReceiveDef
	if receive == nil {
		t.Fatal("Expected receive definition")
	}

	if receive.Name != "process" {
		t.Errorf("Expected method name 'process', got '%s'", receive.Name)
	}

	if len(receive.Body.Statements) != 3 {
		t.Fatalf("Expected 3 body statements, got %d", len(receive.Body.Statements))
	}

	// Test complex object literal with expressions
	setStmt := receive.Body.Statements[0].SetStatement
	if setStmt == nil {
		t.Fatal("Expected set statement")
	}
	if setStmt.Variable != "user" {
		t.Errorf("Expected variable 'user', got '%s'", setStmt.Variable)
	}

	// Test if statement with complex condition
	ifStmt := receive.Body.Statements[1].IfStatement
	if ifStmt == nil {
		t.Fatal("Expected if statement")
	}

	// Test return statement
	returnStmt := receive.Body.Statements[2].ReturnStatement
	if returnStmt == nil {
		t.Fatal("Expected return statement")
	}
}

// Benchmark expression parsing performance
func BenchmarkParser_ExpressionParsing(b *testing.B) {
	code := `receive complex {} -> object {
		set result = {
			sum: a + b + c,
			comparison: x > y && z != null,
			chained: obj.field.subfield
		}
		
		if result.sum > threshold {
			result.status = "high"
		} else {
			result.status = "normal"
		}
		
		return result
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse("benchmark", strings.NewReader(code))
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
