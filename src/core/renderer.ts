import { RelayAST, RelayStatement } from './parser'

export interface RendererContext {
  variables: Record<string, any>
  components: Record<string, any>
}

export class RelayRenderer {
  private context: RendererContext

  constructor() {
    this.context = {
      variables: {},
      components: this.getBuiltInComponents()
    }
  }

  render(ast: RelayAST): JSX.Element {
    return this.renderStatements(ast.body)
  }

  private renderStatements(statements: RelayStatement[]): JSX.Element {
    const elements = statements.map((stmt, index) => {
      switch (stmt.type) {
        case 'relay_block':
          return this.renderRelayBlock(stmt, index)
        case 'show':
          return this.renderShow(stmt, index)
        case 'set':
          this.renderSet(stmt)
          return null
        case 'markdown':
          return this.renderMarkdown(stmt, index)
        default:
          return null
      }
    }).filter(Boolean)

    return React.createElement('div', {}, ...elements)
  }

  private renderRelayBlock(stmt: any, key: number): JSX.Element {
    return React.createElement('div', { key }, 
      this.renderStatements(stmt.statements)
    )
  }

  private renderShow(stmt: any, key: number): JSX.Element {
    const Component = this.context.components[stmt.component]
    if (!Component) {
      throw new Error(`Unknown component: ${stmt.component}`)
    }

    const props = stmt.props.reduce((acc: any, prop: any) => {
      acc[prop.name] = prop.value
      return acc
    }, { key })

    return React.createElement(Component, props)
  }

  private renderSet(stmt: any): void {
    this.context.variables[stmt.name] = stmt.value
  }

  private renderMarkdown(stmt: any, key: number): JSX.Element {
    // Simple markdown rendering - in real app would use proper markdown parser
    return React.createElement('div', { 
      key, 
      dangerouslySetInnerHTML: { __html: stmt.content } 
    })
  }

  private getBuiltInComponents() {
    return {
      heading: ({ text }: { text: string }) => 
        React.createElement('h1', { className: 'text-4xl font-bold mb-4' }, text),
      
      paragraph: ({ text }: { text: string }) => 
        React.createElement('p', { className: 'mb-4' }, text),
      
      button: ({ text }: { text: string }) => 
        React.createElement('button', { 
          className: 'px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600' 
        }, text),
      
      card: ({ children }: { children: React.ReactNode }) => 
        React.createElement('div', { 
          className: 'bg-white rounded-lg shadow-lg p-6 mb-4' 
        }, children),
      
      container: ({ children }: { children: React.ReactNode }) => 
        React.createElement('div', { 
          className: 'max-w-4xl mx-auto p-8' 
        }, children),
      
      grid: ({ children }: { children: React.ReactNode }) => 
        React.createElement('div', { 
          className: 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6' 
        }, children),
      
      column: ({ children }: { children: React.ReactNode }) => 
        React.createElement('div', { 
          className: 'flex flex-col' 
        }, children)
    }
  }
}

// Note: In a real implementation, we'd import React properly
// This is a placeholder for the structure
declare const React: any 