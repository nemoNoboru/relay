# Relay Documentation

Welcome to the Relay language documentation! This directory contains comprehensive documentation for the Relay programming language and its development environment.

## 📚 **Documentation Structure**

### **Language & Vision**
- **[Vision](vision.md)** - The philosophy and goals behind Relay
- **[Roadmap](roadmap.md)** - Development roadmap and future plans

### **Grammar & Language Design**
- **[Grammar](grammar/)** - EBNF grammar definitions
  - `relay-grammar-final.ebnf` - Complete grammar specification
  - `relay-grammar-concise.ebnf` - Simplified grammar overview  
  - `relay-grammar-functions.ebnf` - Function-focused grammar variant

### **Parser Implementation**
- **[Parser](parser/)** - Parser implementation documentation
  - `relay-grammar-implementation.md` - Implementation guide with code examples
  - `relay-implementation-final.md` - Final implementation documentation
  - `parser-summary.md` - Parser implementation summary and test results
  - `parser-integration-summary.md` - Monaco editor integration details

### **Runtime Environment**
- **[Runtime](runtime.md)** - Runtime system overview
- **[Runtime Details](runtime/)** - Detailed runtime documentation
  - `networking-js.md` - Network operations in JavaScript
  - `runtime-js.md` - JavaScript runtime implementation

## 🎯 **Quick Navigation**

### **Getting Started**
1. Start with [Vision](vision.md) to understand Relay's philosophy
2. Review the [Grammar](grammar/relay-grammar-concise.ebnf) for syntax overview
3. Check the [Parser Summary](parser/parser-summary.md) for current implementation status

### **For Developers**
- **Language Design**: See [Grammar](grammar/) directory
- **Parser Development**: See [Parser Implementation](parser/relay-grammar-implementation.md)
- **IDE Integration**: See [Parser Integration](parser/parser-integration-summary.md)
- **Runtime Development**: See [Runtime](runtime.md)

### **Current Status**
- ✅ **Parser**: 82% complete with identifier/string distinction fixed
- ✅ **IDE**: Full Monaco editor integration with syntax highlighting
- ✅ **Grammar**: Complete EBNF specification
- 🚧 **Runtime**: In development

## 🔗 **Related Files**
- Main project README: [../README.md](../README.md)
- Source code: [../src/](../src/)
- Tests: [../src/core/__tests__/](../src/core/__tests__/)

---

*This documentation is maintained alongside the Relay language development. For the latest updates, see the main repository.* 