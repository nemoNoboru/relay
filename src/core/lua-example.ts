// Example usage of Relay Lua Runtime
// This demonstrates how the Lua runtime would work in a browser environment

import { RelayLuaRuntime } from './lua-runtime';

// Example: Book Club App in Lua
const bookClubLuaCode = `
-- Initialize state
local books = relay.state("books", {
    {title = "The Great Gatsby", author = "F. Scott Fitzgerald", rating = 4},
    {title = "To Kill a Mockingbird", author = "Harper Lee", rating = 5},
    {title = "1984", author = "George Orwell", rating = 4}
})

-- Show the app header
relay.show("heading", {text = "My Book Club", level = 1})

-- Show form for adding new books
relay.form("add-book")
relay.show("input", {name = "title", placeholder = "Book title"})
relay.show("input", {name = "author", placeholder = "Author"})
relay.show("button", {text = "Add Book"})

-- Display all books
for i, book in ipairs(books) do
    relay.show("card", {
        title = book.title,
        author = book.author,
        rating = book.rating
    })
end

-- Show statistics
local total_books = #books
local avg_rating = 0
for i, book in ipairs(books) do
    avg_rating = avg_rating + book.rating
end
avg_rating = avg_rating / total_books

relay.show("text", "Total books: " .. total_books)
relay.show("text", "Average rating: " .. string.format("%.1f", avg_rating))

return "Book club app loaded successfully"
`;

// Example: Simple Calculator in Lua
const calculatorLuaCode = `
-- Simple calculator functions
local function add(a, b)
    return relay.math.add(a, b)
end

local function multiply(a, b)
    return relay.math.mul(a, b)
end

-- Test calculations
local result1 = add(10, 5)
local result2 = multiply(3, 4)

relay.show("text", "10 + 5 = " .. result1)
relay.show("text", "3 * 4 = " .. result2)

return {sum = result1, product = result2}
`;

// Example: Data Processing in Lua
const dataProcessingLuaCode = `
-- Process some data
local users = {
    {name = "Alice", age = 25, city = "New York"},
    {name = "Bob", age = 30, city = "San Francisco"},
    {name = "Charlie", age = 35, city = "New York"}
}

-- Filter users by city
local ny_users = {}
for i, user in ipairs(users) do
    if user.city == "New York" then
        table.insert(ny_users, user.name)
    end
end

-- Convert to JSON and back
local json_data = relay.json.stringify(users)
local parsed_data = relay.json.parse(json_data)

relay.show("text", "Users in New York: " .. table.concat(ny_users, ", "))
relay.show("text", "Total users: " .. #users)

return {ny_users = ny_users, total = #users}
`;

// Example usage function (for browser environment)
export async function runLuaExample(exampleName: string): Promise<any> {
    const runtime = new RelayLuaRuntime();
    
    try {
        await runtime.initialize();
        
        let code: string;
        switch (exampleName) {
            case 'book-club':
                code = bookClubLuaCode;
                break;
            case 'calculator':
                code = calculatorLuaCode;
                break;
            case 'data-processing':
                code = dataProcessingLuaCode;
                break;
            default:
                throw new Error(`Unknown example: ${exampleName}`);
        }
        
        const result = await runtime.execute(code);
        console.log(`[LUA] ${exampleName} result:`, result);
        return result;
        
    } catch (error) {
        console.error(`[LUA] Error running ${exampleName}:`, error);
        throw error;
    } finally {
        runtime.close();
    }
}

// Export the example codes for reference
export const luaExamples = {
    'book-club': bookClubLuaCode,
    'calculator': calculatorLuaCode,
    'data-processing': dataProcessingLuaCode
}; 