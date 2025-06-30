# Relay Desktop

**Web Development for Everyone**

Relay Desktop is an innovative development environment that makes building dynamic websites accessible to everyone. Write in simple, English-like syntax and see your changes live as you type.

## Features

- 🚀 **Simple Syntax** - English-like commands that anyone can understand
- 🔥 **Live Preview** - See your changes instantly as you type
- 🎨 **Built-in Components** - Rich library of UI components ready to use
- 📁 **Project Management** - Organized file structure with pages, functions, and data
- 🎯 **Monaco Editor** - Professional code editor with syntax highlighting
- ⚡ **Fast Development** - Built with Vite for lightning-fast development

## Quick Start

### Prerequisites

- [Bun](https://bun.sh/) - Fast JavaScript runtime and package manager

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd relay
```

2. Install dependencies with Bun:
```bash
bun install
```

3. Start the development server:
```bash
bun run electron:dev
```

## Development Scripts

- `bun run dev` - Start Vite development server
- `bun run electron:dev` - Start Electron app in development mode
- `bun run build` - Build the application for production
- `bun run electron` - Run the built Electron app
- `bun run preview` - Preview the built application

## Project Structure

```
src/
├── app/                 # Relay desktop application
│   ├── components/      # React components
│   ├── types.ts         # TypeScript definitions
│   └── App.tsx          # Main application component
├── core/                # Relay language core
│   ├── parser.ts        # PEG.js parser for Relay syntax
│   └── renderer.ts      # Renders parsed AST to React
├── main.tsx             # React entry point
└── index.css            # Global styles with Tailwind

electron/
├── main.ts              # Electron main process
└── preload.ts           # Electron preload script

public/
└── relay-icon.svg       # Application icon
```

## Relay Language Syntax

### Basic Structure

```relay
relay {
  show heading "Welcome to Relay"
  show paragraph "This is a simple example"
}
```

### Variables and Data

```relay
relay {
  set greeting "Hello, World!"
  show heading (get greeting)
}
```

### Components

```relay
relay {
  show card
    show heading "My Card"
    show paragraph "Card content goes here"
    show button "Click me"
}
```

### Loops and Logic

```relay
relay {
  set items ["Item 1", "Item 2", "Item 3"]
  
  for item in items
    show list-item (get item)
}
```

## Built-in Components

- **Layout**: `container`, `grid`, `column`, `card`
- **Typography**: `heading`, `paragraph`
- **Forms**: `input-text`, `button`, `select`, `checkbox`
- **Data**: `list-item`, `table-row`
- **Media**: `image`

## Technology Stack

- **Electron** - Desktop application framework
- **React** - UI library
- **TypeScript** - Type-safe JavaScript
- **Vite** - Build tool and development server
- **Tailwind CSS** - Utility-first CSS framework
- **Monaco Editor** - Code editor (VS Code's editor)
- **Bun** - JavaScript runtime and package manager

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

See [roadmap.md](roadmap.md) for detailed development plans and milestones.

## Vision

Read our [vision.md](vision.md) to understand the mission behind Relay and why we're building it.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Community

- [Discord Community](https://discord.gg/relay) - Join our community
- [GitHub Discussions](https://github.com/relay/relay/discussions) - Ask questions and share ideas
- [Twitter](https://twitter.com/relayapp) - Follow for updates

---

**Made with ❤️ for everyone who wants to build for the web** 