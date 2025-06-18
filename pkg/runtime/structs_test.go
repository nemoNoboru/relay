package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectCreation(t *testing.T) {
	t.Run("Simple object literal", func(t *testing.T) {
		result := evalCode(t, `{name: "John", age: 30}`)
		require.Equal(t, ValueTypeObject, result.Type)
		require.Contains(t, result.Object, "name")
		require.Contains(t, result.Object, "age")
		require.Equal(t, "John", result.Object["name"].Str)
		require.Equal(t, 30.0, result.Object["age"].Number)
	})

	t.Run("Empty object", func(t *testing.T) {
		result := evalCode(t, `{}`)
		require.Equal(t, ValueTypeObject, result.Type)
		require.Len(t, result.Object, 0)
	})

	t.Run("Object with variables", func(t *testing.T) {
		result := evalCode(t, `
		set name = "Alice"
		set age = 25
		{name: name, age: age}`)
		require.Equal(t, ValueTypeObject, result.Type)
		require.Equal(t, "Alice", result.Object["name"].Str)
		require.Equal(t, 25.0, result.Object["age"].Number)
	})

	t.Run("Object with mixed types", func(t *testing.T) {
		result := evalCode(t, `{
			str: "hello",
			num: 42,
			bool: true,
			arr: [1, 2, 3],
			obj: {nested: "value"}
		}`)
		require.Equal(t, ValueTypeObject, result.Type)
		require.Equal(t, ValueTypeString, result.Object["str"].Type)
		require.Equal(t, ValueTypeNumber, result.Object["num"].Type)
		require.Equal(t, ValueTypeBool, result.Object["bool"].Type)
		require.Equal(t, ValueTypeArray, result.Object["arr"].Type)
		require.Equal(t, ValueTypeObject, result.Object["obj"].Type)
	})
}

func TestObjectFieldAccess(t *testing.T) {
	t.Run("Basic field access", func(t *testing.T) {
		result := evalCode(t, `
		set person = {name: "Bob", age: 35}
		person.name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Bob", result.Str)
	})

	t.Run("Number field access", func(t *testing.T) {
		result := evalCode(t, `
		set person = {name: "Bob", age: 35}
		person.age`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 35.0, result.Number)
	})

	t.Run("Nested object access", func(t *testing.T) {
		result := evalCode(t, `
		set data = {
			user: {
				profile: {
					name: "Carol"
				}
			}
		}
		data.user.profile.name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Carol", result.Str)
	})

	t.Run("Array in object access", func(t *testing.T) {
		result := evalCode(t, `
		set data = {numbers: [10, 20, 30]}
		data.numbers.get(1)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 20.0, result.Number)
	})
}

func TestObjectFieldModification(t *testing.T) {
	t.Run("Modify existing field", func(t *testing.T) {
		result := evalCode(t, `
		set person = {name: "David", age: 40}
		person.age = 41
		person.age`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 41.0, result.Number)
	})

	t.Run("Add new field", func(t *testing.T) {
		result := evalCode(t, `
		set person = {name: "Eve"}
		person.age = 28
		person.age`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 28.0, result.Number)
	})

	t.Run("Modify nested field", func(t *testing.T) {
		result := evalCode(t, `
		set data = {user: {profile: {name: "Frank"}}}
		data.user.profile.name = "Franklin"
		data.user.profile.name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Franklin", result.Str)
	})
}

func TestStructDefinition(t *testing.T) {
	t.Run("Simple struct definition", func(t *testing.T) {
		result := evalCode(t, `
		struct Person {
			name: string,
			age: number
		}
		Person {name: "Grace", age: 30}`)
		require.Equal(t, ValueTypeStruct, result.Type)
		require.Equal(t, "Person", result.Struct.Name)
		require.Contains(t, result.Struct.Fields, "name")
		require.Contains(t, result.Struct.Fields, "age")
		require.Equal(t, "Grace", result.Struct.Fields["name"].Str)
		require.Equal(t, 30.0, result.Struct.Fields["age"].Number)
	})

	t.Run("Struct with array field", func(t *testing.T) {
		result := evalCode(t, `
		struct Container {
			items: array,
			count: number
		}
		Container {items: [1, 2, 3], count: 3}`)
		require.Equal(t, ValueTypeStruct, result.Type)
		require.Equal(t, "Container", result.Struct.Name)
		require.Equal(t, ValueTypeArray, result.Struct.Fields["items"].Type)
		require.Len(t, result.Struct.Fields["items"].Array, 3)
	})

	t.Run("Nested struct", func(t *testing.T) {
		result := evalCode(t, `
		struct Address {
			street: string,
			city: string
		}
		struct Person {
			name: string,
			address: Address
		}
		Person {
			name: "Henry",
			address: Address {street: "Main St", city: "NYC"}
		}`)
		require.Equal(t, ValueTypeStruct, result.Type)
		require.Equal(t, "Person", result.Struct.Name)
		require.Equal(t, ValueTypeStruct, result.Struct.Fields["address"].Type)
		require.Equal(t, "Address", result.Struct.Fields["address"].Struct.Name)
	})
}

func TestStructFieldAccess(t *testing.T) {
	t.Run("Access struct field", func(t *testing.T) {
		result := evalCode(t, `
		struct Person {
			name: string,
			age: number
		}
		set person = Person {name: "Ivy", age: 25}
		person.name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Ivy", result.Str)
	})

	t.Run("Access nested struct field", func(t *testing.T) {
		result := evalCode(t, `
		struct Address {
			street: string,
			city: string
		}
		struct Person {
			name: string,
			address: Address
		}
		set person = Person {
			name: "Jack",
			address: Address {street: "Oak Ave", city: "LA"}
		}
		person.address.city`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "LA", result.Str)
	})

	t.Run("Access array in struct", func(t *testing.T) {
		result := evalCode(t, `
		struct Team {
			name: string,
			scores: array
		}
		set team = Team {name: "Lions", scores: [10, 15, 20]}
		team.scores.get(1)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 15.0, result.Number)
	})
}

func TestStructsWithFunctions(t *testing.T) {
	t.Run("Function returning struct", func(t *testing.T) {
		result := evalCode(t, `
		struct Point {
			x: number,
			y: number
		}
		fn create_point(x: number, y: number) -> Point {
			Point {x: x, y: y}
		}
		set p = create_point(3, 4)
		p.x`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 3.0, result.Number)
	})

	t.Run("Function taking struct parameter", func(t *testing.T) {
		result := evalCode(t, `
		struct Point {
			x: number,
			y: number
		}
		fn distance_from_origin(p: Point) -> number {
			p.x * p.x + p.y * p.y
		}
		set point = Point {x: 3, y: 4}
		distance_from_origin(point)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 25.0, result.Number) // 3^2 + 4^2 = 9 + 16 = 25
	})

	t.Run("Function modifying struct", func(t *testing.T) {
		result := evalCode(t, `
		struct Counter {
			value: number
		}
		fn increment(c: Counter) -> number {
			c.value = c.value + 1
			c.value
		}
		set counter = Counter {value: 5}
		increment(counter)`)
		require.Equal(t, ValueTypeNumber, result.Type)
		require.Equal(t, 6.0, result.Number)
	})
}

func TestComplexObjectScenarios(t *testing.T) {
	t.Run("Array of objects", func(t *testing.T) {
		result := evalCode(t, `
		set users = [
			{name: "Alice", age: 30},
			{name: "Bob", age: 25},
			{name: "Charlie", age: 35}
		]
		users.get(1).name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Bob", result.Str)
	})

	t.Run("Array of structs", func(t *testing.T) {
		result := evalCode(t, `
		struct Person {
			name: string,
			age: number
		}
		set people = [
			Person {name: "Dana", age: 28},
			Person {name: "Erik", age: 32}
		]
		people.get(0).name`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Dana", result.Str)
	})

	t.Run("Object with function field", func(t *testing.T) {
		result := evalCode(t, `
		fn greet(name: string) -> string {
			"Hello, " + name
		}
		set obj = {
			message: "Welcome",
			greeter: greet
		}
		obj.greeter("World")`)
		require.Equal(t, ValueTypeString, result.Type)
		require.Equal(t, "Hello, World", result.Str)
	})
}

func TestStructAndObjectErrors(t *testing.T) {
	t.Run("Access undefined field", func(t *testing.T) {
		err := evalCodeError(t, `
		set obj = {name: "Test"}
		obj.undefined_field`)
		require.Error(t, err)
	})

	t.Run("Access field on non-object", func(t *testing.T) {
		err := evalCodeError(t, `
		set num = 42
		num.field`)
		require.Error(t, err)
	})

	t.Run("Undefined struct type", func(t *testing.T) {
		err := evalCodeError(t, `UndefinedStruct {field: "value"}`)
		require.Error(t, err)
	})

	t.Run("Struct with missing required field", func(t *testing.T) {
		err := evalCodeError(t, `
		struct Person {
			name: string,
			age: number
		}
		Person {name: "Test"}`) // Missing age field
		require.Error(t, err)
	})
}
