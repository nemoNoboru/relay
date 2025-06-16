package parser

import (
	"strings"
	"testing"
)

func TestForStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "basic for loop",
			input: `receive test {} -> object {
				for item in items {
					set x = item.value
				}
			}`,
		},
		{
			name: "for loop with field access",
			input: `receive test {} -> object {
				for user in users.active {
					set total = total + user.score
				}
			}`,
		},
		{
			name: "nested for loops",
			input: `receive test {} -> object {
				for group in groups {
					for member in group.members {
						set processed = member.name
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("Parse() error = %v", err)
			}
		})
	}
}

func TestTryStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "basic try-catch",
			input: `receive test {} -> object {
				try {
					set result = risky_operation()
				} catch error {
					set result = "failed"
				}
			}`,
		},
		{
			name: "try-catch without variable",
			input: `receive test {} -> object {
				try {
					set user = user_data.name
				} catch {
					return {error: "user not found"}
				}
			}`,
		},
		{
			name: "nested try-catch",
			input: `receive test {} -> object {
				try {
					try {
						set data = input.json
					} catch parse_error {
						throw {error: "invalid json"}
					}
				} catch main_error {
					return {error: "processing failed"}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("Parse() error = %v", err)
			}
		})
	}
}

func TestDispatchStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "basic dispatch",
			input: `receive test {} -> object {
				dispatch activity.type {
					"Create" -> {
						return activity.title
					}
					"Delete" -> {
						return activity.id
					}
				}
			}`,
		},
		{
			name: "dispatch with numbers",
			input: `receive test {} -> object {
				dispatch status_code {
					200 -> {
						return "success"
					}
					404 -> {
						return "not found"
					}
					500 -> {
						return "server error"
					}
				}
			}`,
		},
		{
			name: "dispatch with boolean",
			input: `receive test {} -> object {
				dispatch is_admin {
					true -> {
						return admin_dashboard()
					}
					false -> {
						return user_dashboard()
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("Parse() error = %v", err)
			}
		})
	}
}

func TestComplexStatementCombinations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "for loop with try-catch",
			input: `receive test {} -> object {
				for user in users {
					try {
						set result = send "user_service" validate_user {id: user.id}
						if result.valid {
							processed = processed.length + 1
						}
					} catch error {
						errors = errors.length + 1
					}
				}
			}`,
		},
		{
			name: "dispatch with send statements",
			input: `receive handle_activity {} -> object {
				dispatch activity.type {
					"Create" -> {
						set post = send "blog_service" create_post {
							title: activity.object.title,
							content: activity.object.content
						}
						return post
					}
					"Follow" -> {
						set result = send "user_service" add_follower {
							username: activity.object.username,
							follower: activity.actor
						}
						return result
					}
				}
			}`,
		},
		{
			name: "nested control flow",
			input: `receive complex_operation {} -> object {
				set results = []
				
				for batch in data_batches {
					try {
						for item in batch.items {
							dispatch item.type {
								"process" -> {
									set processed = send "processor" handle {data: item}
									results = results.length + 1
								}
								"validate" -> {
									if item.valid {
										results = results.length + 1
									}
								}
							}
						}
					} catch batch_error {
						set error_log = send "logger" log_error {
							batch: batch.id,
							error: batch_error
						}
					}
				}
				
				return {results: results, total: results.length}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("test.relay", strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("Parse() error = %v", err)
			}
		})
	}
}

func TestStatementParsing(t *testing.T) {
	input := `receive comprehensive_test {} -> object {
		// Set and return statements
		set user = {name: "John", age: 30}
		
		// If-else statement
		if user.age > 18 {
			set status = "adult"
		} else {
			set status = "minor"
		}
		
		// For loop
		for item in items {
			set total = total + item.value
		}
		
		// Try-catch
		try {
			set result = risky_data.value
		} catch error {
			throw {error: "operation failed"}
		}
		
		// Dispatch
		dispatch user.role {
			"admin" -> {
				return admin_data.dashboard
			}
			"user" -> {
				return user_data.profile
			}
		}
		
		// Send statement
		set posts = send "blog_service" get_posts {limit: 10}
		
		// Assignments
		user.score += 100
		user.name = "Jane"
		
		return user
	}`

	program, err := Parse("test.relay", strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(program.Statements) == 0 {
		t.Fatal("Expected at least one statement")
	}

	receiveDef := program.Statements[0].ReceiveDef
	if receiveDef == nil {
		t.Fatal("Expected receive definition")
	}

	statements := receiveDef.Body.Statements
	if len(statements) == 0 {
		t.Fatal("Expected statements in receive body")
	}

	// Check that we have various statement types
	foundTypes := make(map[string]bool)

	var checkStatements func([]*BlockStatement)
	checkStatements = func(stmts []*BlockStatement) {
		for _, stmt := range stmts {
			if stmt.SetStatement != nil {
				foundTypes["set"] = true
			}
			if stmt.IfStatement != nil {
				foundTypes["if"] = true
				// Check if and else blocks
				if stmt.IfStatement.ThenBlock != nil {
					checkStatements(stmt.IfStatement.ThenBlock.Statements)
				}
				if stmt.IfStatement.ElseBlock != nil {
					checkStatements(stmt.IfStatement.ElseBlock.Statements)
				}
			}
			if stmt.ForStatement != nil {
				foundTypes["for"] = true
				if stmt.ForStatement.Body != nil {
					checkStatements(stmt.ForStatement.Body.Statements)
				}
			}
			if stmt.TryStatement != nil {
				foundTypes["try"] = true
				// Check try and catch blocks
				if stmt.TryStatement.TryBlock != nil {
					checkStatements(stmt.TryStatement.TryBlock.Statements)
				}
				if stmt.TryStatement.CatchBlock != nil {
					checkStatements(stmt.TryStatement.CatchBlock.Statements)
				}
			}
			if stmt.DispatchStatement != nil {
				foundTypes["dispatch"] = true
				// Check dispatch cases
				for _, dispatchCase := range stmt.DispatchStatement.Cases {
					if dispatchCase.Body != nil {
						checkStatements(dispatchCase.Body.Statements)
					}
				}
			}
			if stmt.Assignment != nil {
				foundTypes["assignment"] = true
			}
			if stmt.ReturnStatement != nil {
				foundTypes["return"] = true
			}
			if stmt.ThrowStatement != nil {
				foundTypes["throw"] = true
			}
		}
	}

	checkStatements(statements)

	expectedTypes := []string{"set", "if", "for", "try", "dispatch", "assignment", "return", "throw"}
	for _, expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Expected to find %s statement", expectedType)
		}
	}
}

// Benchmark the statement parsing performance
func BenchmarkStatementParsing(b *testing.B) {
	input := `receive benchmark_test {} -> object {
		for user in users {
			try {
				dispatch user.type {
					"admin" -> {
						set result = send "admin_service" process {user: user}
						return result
					}
					"user" -> {
						set result = send "user_service" process {user: user}
						return result
					}
				}
			} catch error {
				throw {error: "processing failed", user: user.id}
			}
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse("benchmark.relay", strings.NewReader(input))
		if err != nil {
			b.Fatalf("Parse() error = %v", err)
		}
	}
}
