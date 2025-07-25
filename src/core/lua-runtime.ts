// Relay Lua Runtime using Wasmoon
// Provides a sandboxed Lua environment with Relay API

import { LuaFactory, LuaEngine } from 'wasmoon';

// Environment for variable and function scoping
export interface Environment {
  parent?: Environment;
  bindings: Map<string, any>;
}

// Renderable component for the show function
export interface RenderableComponent {
  type: 'component';
  name: string;
  props: Record<string, any>;
  children?: RenderableComponent[];
}

// Helper function to create a renderable component
export function createComponent(name: string, props: Record<string, any> = {}, children: RenderableComponent[] = []): RenderableComponent {
  console.log(`[CREATE] Creating component "${name}" with props:`, props)
  const component = {
    type: 'component' as const,
    name,
    props,
    children
  };
  console.log(`[CREATE] Created component:`, component)
  return component;
}

// Main Lua runtime class
export class RelayLuaRuntime {
  private factory: LuaFactory;
  private engine: LuaEngine | null = null;
  private globalEnv: Environment;
  private componentCollection: RenderableComponent[] = [];
  private isEvaluatingChildren: boolean = false;

  constructor() {
    this.factory = new LuaFactory();
    this.globalEnv = this.createEnvironment();
    this.componentCollection = [];
  }

  // Initialize the Lua runtime
  async initialize(): Promise<void> {
    try {
      this.engine = await this.factory.createEngine({
        openStandardLibs: true,
        injectObjects: true
      });
      this.setupRelayAPI();
      console.log('[LUA] Runtime initialized successfully');
    } catch (error) {
      console.error('[LUA] Failed to initialize runtime:', error);
      throw error;
    }
  }

  // Create a new environment
  private createEnvironment(parent?: Environment): Environment {
    return {
      parent,
      bindings: new Map()
    };
  }

  // Set child evaluation mode
  setChildEvaluationMode(isChild: boolean): void {
    this.isEvaluatingChildren = isChild;
  }

  // Add component to collection
  addComponent(component: RenderableComponent): void {
    if (!this.isEvaluatingChildren) {
      this.componentCollection.push(component);
    }
  }

  // Set variable in global environment
  setVariable(name: string, value: any): void {
    this.globalEnv.bindings.set(name, value);
  }

  // Setup Relay API for Lua
  private setupRelayAPI(): void {
    if (!this.engine) {
      throw new Error('Lua runtime not initialized');
    }

    // Initialize event handlers in Lua global scope
    this.engine.global.set('_eventHandlers', {});

    // Expose Relay API to Lua
    this.engine.global.set('relay', {
      // UI Components
      show: (component: string, props?: any) => {
        console.log(`[LUA] show() called with component: "${component}", props:`, props)
        console.log(`[LUA] props type:`, typeof props)
        console.log(`[LUA] props keys:`, props ? Object.keys(props) : 'undefined')
        const componentObj = createComponent(component, props || {});
        console.log(`[LUA] Created component object:`, componentObj)
        this.addComponent(componentObj);
        console.log(`[LUA] Added component to collection. Total components:`, this.componentCollection.length);
        return componentObj;
      },

      // State Management
      state: (name: string, initialValue?: any) => {
        // If no initialValue provided, just return the current value
        if (initialValue === undefined) {
          return this.globalEnv.bindings.get(name);
        }
        
        // If the state doesn't exist, initialize it
        if (!this.globalEnv.bindings.has(name)) {
          this.globalEnv.bindings.set(name, initialValue);
          console.log(`[LUA] Initialized state: ${name} =`, initialValue);
        } else {
          // Update existing state
          this.globalEnv.bindings.set(name, initialValue);
          console.log(`[LUA] Updated state: ${name} =`, initialValue);
        }
        
        return this.globalEnv.bindings.get(name);
      },

      // HTTP Requests
      fetch: async (url: string, options?: any) => {
        try {
          const response = await fetch(url, options);
          const data = await response.json();
          console.log(`[LUA] Fetch successful: ${url}`, data);
          return data;
        } catch (error) {
          console.error(`[LUA] Fetch failed: ${url}`, error);
          throw error;
        }
      },

      // Form Handling
      form: (name: string, handler?: any) => {
        const formComponent = createComponent('form', { name, onSubmit: name });
        this.addComponent(formComponent);
        console.log(`[LUA] Created form: ${name}`);
        return formComponent;
      },

      // Event Handling
      when: (eventName: string, handler?: any) => {
        const eventHandler = {
          type: 'event_handler',
          eventName,
          handler,
          env: this.globalEnv
        };

        if (!this.globalEnv.bindings.has('_eventHandlers')) {
          this.globalEnv.bindings.set('_eventHandlers', new Map());
        }
        const eventHandlers = this.globalEnv.bindings.get('_eventHandlers');
        eventHandlers.set(eventName, eventHandler);

        // Also store in Lua global scope
        if (this.engine) {
          const luaEventHandlers = this.engine.global.get('_eventHandlers') || {};
          luaEventHandlers[eventName] = eventHandler;
          this.engine.global.set('_eventHandlers', luaEventHandlers);
        }

        console.log(`[LUA] Registered event handler: ${eventName}`);
        return eventHandler;
      },

      // Storage
      storage: {
        read: (path: string) => {
          // TODO: Implement file reading from Relay storage
          console.log(`[LUA] Reading from storage: ${path}`);
          return null;
        },
        write: (path: string, data: any) => {
          // TODO: Implement file writing to Relay storage
          console.log(`[LUA] Writing to storage: ${path}`, data);
          return true;
        }
      },

      // Utility functions
      json: {
        parse: (text: string) => JSON.parse(text),
        stringify: (obj: any) => JSON.stringify(obj)
      },

      // Math operations
      math: {
        add: (a: number, b: number) => a + b,
        sub: (a: number, b: number) => a - b,
        mul: (a: number, b: number) => a * b,
        div: (a: number, b: number) => a / b,
        mod: (a: number, b: number) => a % b
      },

      // String operations
      string: {
        concat: (...args: any[]) => args.join(''),
        length: (str: string) => str.length,
        upper: (str: string) => str.toUpperCase(),
        lower: (str: string) => str.toLowerCase()
      }
    });
  }

  // Execute Lua code
  async execute(code: string): Promise<any> {
    if (!this.engine) {
      throw new Error('Lua runtime not initialized');
    }

    try {
      // Reset component collection for each execution
      this.componentCollection = [];
      
      const result = await this.engine.doString(code);
      
      // Extract event handlers from global environment
      const eventHandlers = this.globalEnv.bindings.get('_eventHandlers') || new Map();
      
      // If we have collected components, return them as a collection
      if (this.componentCollection.length > 0) {
        console.log(`[LUA] Returning component collection with ${this.componentCollection.length} components:`, this.componentCollection);
        return {
          type: 'component_collection',
          components: this.componentCollection,
          eventHandlers: eventHandlers
        };
      }
      
      // Return result with event handlers if any exist
      if (eventHandlers.size > 0) {
        return {
          type: 'result_with_handlers',
          value: result,
          eventHandlers: eventHandlers
        };
      }
      
      return result;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(`Lua execution error: ${message}`);
    }
  }

  // Execute a complete program (multiple code blocks)
  async executeProgram(codeBlocks: string[]): Promise<any> {
    let lastResult = null;
    
    for (const codeBlock of codeBlocks) {
      lastResult = await this.execute(codeBlock);
    }
    
    return lastResult;
  }

  // Clean up resources
  close(): void {
    if (this.engine) {
      // Note: Wasmoon doesn't have a close method on engine
      // The engine will be garbage collected
      this.engine = null;
    }
  }

  public async executeString(code: string): Promise<any> {
    if (!this.engine) throw new Error('Lua runtime not initialized');
    return this.engine.doString(code);
  }
}

// Convenience function to run Lua code
export async function runLua(code: string): Promise<any> {
  const runtime = new RelayLuaRuntime();
  try {
    await runtime.initialize();
    return await runtime.execute(code);
  } finally {
    runtime.close();
  }
} 