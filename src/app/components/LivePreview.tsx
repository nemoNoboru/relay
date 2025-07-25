import React, { useState, useEffect, useRef } from 'react'
import { RenderableComponent } from '../../core/lua-runtime'
import { twMerge } from 'tailwind-merge'

interface LivePreviewProps {
  // Instead of content, we receive the Lua runtime result and event handlers
  interpreterResult: any
  eventHandlers: Map<string, any>
  error?: string
  isLoading?: boolean
  // Callback to send events back to parent
  onEventTriggered: (eventName: string, eventData: any) => void
  // Form state management props
  formValues: Record<string, string>
  onFormValueUpdate: (name: string, value: string) => void
}

type ViewMode = 'preview' | 'raw' | 'execute'

export default function LivePreview({ 
  interpreterResult, 
  eventHandlers, 
  error, 
  isLoading = false,
  onEventTriggered,
  formValues,
  onFormValueUpdate
}: LivePreviewProps) {
  console.log(`[LIVE_PREVIEW] Rendering with formValues:`, formValues)
  console.log(`[LIVE_PREVIEW] interpreterResult:`, interpreterResult)
  const [renderedContent, setRenderedContent] = useState<JSX.Element | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('preview')
  
  // Use ref to get current formValues in event handlers
  const formValuesRef = useRef(formValues)
  formValuesRef.current = formValues

  useEffect(() => {
    if (interpreterResult) {
      const rendered = renderFromLuaResult(interpreterResult)
      setRenderedContent(rendered)
    } else if (error) {
      setRenderedContent(renderError(error))
    } else {
      setRenderedContent(renderEmptyState())
    }
  }, [interpreterResult, error])

  // Function to execute event handlers - now sends message to parent
  const executeEventHandler = (eventName: string, eventData: any) => {
    console.log(`[EVENT] LivePreview sending event: ${eventName}`, eventData)
    onEventTriggered(eventName, eventData)
  }

  // Handle input value changes
  const updateFormValue = (name: string, value: string) => {
    onFormValueUpdate(name, value)
  }

  // Render error state
  const renderError = (errorMessage: string): JSX.Element => {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <div className="flex items-center">
          <svg className="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 className="text-red-800 font-medium">Error</h3>
        </div>
        <p className="text-red-700 mt-2 text-sm">{errorMessage}</p>
      </div>
    )
  }

  // Render empty state
  const renderEmptyState = (): JSX.Element => {
    return (
      <div className="p-8 bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 mx-auto mb-4 bg-gray-200 rounded-full flex items-center justify-center">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Content</h3>
          <p className="text-gray-500">Start writing Lua code to see the preview</p>
        </div>
      </div>
    )
  }

  // New function to render based on Lua runtime result
  const renderFromLuaResult = (result: any): JSX.Element => {
    console.log('🔍 renderFromLuaResult called with:', result)
    console.log('🔍 Result type:', typeof result)
    console.log('🔍 Result keys:', result && typeof result === 'object' ? Object.keys(result) : 'N/A')
    
    // Handle component collection from Lua runtime
    if (result && typeof result === 'object' && result.components && Array.isArray(result.components)) {
      console.log('🎯 Rendering component collection from Lua:', result.components)
      console.log('🎯 Component collection length:', result.components.length)
      return renderComponentCollection(result.components)
    }
    
    // Handle single component
    if (result && typeof result === 'object' && result.type === 'component') {
      console.log('🎯 Rendering single component from Lua:', result)
      return renderComponent(result as RenderableComponent)
    }
    
    // Handle arrays of components (direct array format)
    if (Array.isArray(result)) {
      // Check if it's an array of components
      const firstItem = result[0]
      if (firstItem && typeof firstItem === 'object' && firstItem.type === 'component') {
        console.log('🎯 Rendering array of components from Lua:', result)
        return renderComponentCollection(result as RenderableComponent[])
      }
    }
    
    // Handle primitive values
    if (typeof result === 'string' || typeof result === 'number' || typeof result === 'boolean') {
      return (
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <pre className="text-sm font-mono">{String(result)}</pre>
        </div>
      )
    }
    
    // Handle objects (display as JSON)
    if (typeof result === 'object' && result !== null) {
      return (
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <pre className="text-sm font-mono overflow-auto">
            {JSON.stringify(result, null, 2)}
          </pre>
        </div>
      )
    }
    
    // Fallback
    return renderEmptyState()
  }

  const renderComponentCollection = (components: RenderableComponent[]): JSX.Element => {
    console.log('🎨 renderComponentCollection called with:', components.length, 'components')
    return (
      <div className="space-y-4 p-4">
        {components.map((component: RenderableComponent, index: number) => {
          console.log(`🎨 Rendering component ${index}:`, component.name, component.props)
          // Use a stable key for input components to prevent re-mounting
          const key = component.name === 'input' && component.props.name 
            ? `input-${component.props.name}` 
            : `${component.name}-${index}`
          return (
            <div key={key}>
              {renderComponent(component)}
            </div>
          )
        })}
      </div>
    )
  }

  const mergeClasses = (defaultClasses: string, customClasses?: string): string => {
    if (!customClasses) return defaultClasses
    return twMerge(defaultClasses, customClasses)
  }

  const renderComponentContent = (component: RenderableComponent): JSX.Element => {
    const { props } = component
    
    // Handle text content
    if (props.text) {
      return <span>{props.text}</span>
    }
    
    // Handle children
    if (component.children && component.children.length > 0) {
      return (
        <>
          {component.children.map((child: RenderableComponent, index: number) => (
            <div key={index}>
              {renderComponent(child)}
            </div>
          ))}
        </>
      )
    }
    
    // Handle content prop
    if (props.content) {
      return <span>{props.content}</span>
    }
    
    return <></>
  }

  const renderComponent = (component: RenderableComponent): JSX.Element => {
    const { name, props } = component
    console.log(`[RENDER] Rendering component: ${name}`, props)
    
    // Common props
    const className = mergeClasses('', props.className || props.class)
    const id = props.id
    const style = props.style || {}
    
    // Handle different component types
    switch (name) {
      case 'heading':
        const level = props.level || 1
        const HeadingTag = `h${level}` as keyof JSX.IntrinsicElements
        return (
          <HeadingTag 
            className={mergeClasses('font-bold text-gray-900', className)}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </HeadingTag>
        )
        
      case 'paragraph':
        return (
          <p 
            className={mergeClasses('text-gray-700 leading-relaxed', className)}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </p>
        )
        
      case 'card':
        return (
          <div 
            className={mergeClasses(
              'bg-white border border-gray-200 rounded-lg shadow-sm p-6',
              className
            )}
            id={id}
            style={style}
          >
            {props.title && (
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {props.title}
              </h3>
            )}
            {renderComponentContent(component)}
          </div>
        )
        
      case 'button':
        const handleClick = () => {
          if (props.onClick) {
            // Pass the current form values to the event handler
            executeEventHandler(props.onClick, formValuesRef.current)
          }
        }
        
        return (
          <button
            className={mergeClasses(
              'px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white font-medium rounded-lg transition-colors',
              className
            )}
            id={id}
            style={style}
            onClick={handleClick}
            disabled={props.disabled}
          >
            {renderComponentContent(component)}
          </button>
        )
        
      case 'input':
        const inputName = props.name || 'input'
        const inputValue = formValuesRef.current[inputName] || props.value || ''
        
        const handleInputChangeEvent = (e: React.ChangeEvent<HTMLInputElement>) => {
          const value = e.target.value
          updateFormValue(inputName, value)
          
          if (props.onChange) {
            executeEventHandler(props.onChange, { value, name: inputName })
          }
        }
        
        const handleInputFocus = (e: React.FocusEvent<HTMLInputElement>) => {
          console.log(`[INPUT] Focus event for ${inputName}`)
        }
        
        const handleInputBlur = (e: React.FocusEvent<HTMLInputElement>) => {
          console.log(`[INPUT] Blur event for ${inputName}`)
        }
        
        const inputElement = (
          <input
            type={props.type || 'text'}
            placeholder={props.placeholder}
            defaultValue={formValuesRef.current[inputName] || ''}
            className={mergeClasses(
              'w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500',
              className
            )}
            id={id}
            style={style}
            onChange={handleInputChangeEvent}
            onFocus={handleInputFocus}
            onBlur={handleInputBlur}
            name={inputName}
            autoComplete={props.autoComplete || 'off'}
          />
        )
        return inputElement
        
      case 'row':
        return (
          <div 
            className={mergeClasses('flex flex-row gap-4', className)}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </div>
        )
        
      case 'col':
        return (
          <div 
            className={mergeClasses('flex flex-col gap-4', className)}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </div>
        )
        
      case 'grid':
        const columns = props.columns || 3
        return (
          <div 
            className={mergeClasses(
              `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-${columns} gap-4`,
              className
            )}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </div>
        )
        
      case 'form':
        const handleSubmit = (e: React.FormEvent) => {
          e.preventDefault()
          if (props.onSubmit) {
            // Use current form values instead of FormData
            executeEventHandler(props.onSubmit, formValues)
          }
        }
        
        return (
          <form 
            className={mergeClasses('space-y-4', className)}
            id={id}
            style={style}
            onSubmit={handleSubmit}
          >
            {renderComponentContent(component)}
          </form>
        )
        
      case 'list':
        const items = props.items || []
        return (
          <ul className={mergeClasses('list-disc list-inside space-y-2', className)}>
            {items.map((item: any, index: number) => (
              <li key={index} className="text-gray-700">
                {typeof item === 'string' ? item : item.text || item.content || JSON.stringify(item)}
              </li>
            ))}
          </ul>
        )
        
      case 'table':
        const tableData = props.data || []
        const headers = props.headers || []
        
        return (
          <div className="overflow-x-auto">
            <table className={mergeClasses('min-w-full border border-gray-200', className)}>
              {headers.length > 0 && (
                <thead>
                  <tr>
                    {headers.map((header: string, index: number) => (
                      <th key={index} className="border border-gray-200 px-4 py-2 bg-gray-50 font-medium">
                        {header}
                      </th>
                    ))}
                  </tr>
                </thead>
              )}
              <tbody>
                {tableData.map((row: any[], rowIndex: number) => (
                  <tr key={rowIndex}>
                    {row.map((cell: any, cellIndex: number) => (
                      <td key={cellIndex} className="border border-gray-200 px-4 py-2">
                        {typeof cell === 'string' ? cell : JSON.stringify(cell)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
        
      default:
        // Generic div for unknown components
        return (
          <div 
            className={mergeClasses('', className)}
            id={id}
            style={style}
          >
            {renderComponentContent(component)}
          </div>
        )
    }
  }

  return (
    <div className="h-full flex flex-col">
      {/* Preview Header */}
      <div className="border-b border-gray-200 bg-gray-50 px-4 py-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-gray-700">Live Preview</h2>
          <div className="flex items-center space-x-2">
            <button 
              className={`px-2 py-1 text-xs rounded ${
                viewMode === 'preview' 
                  ? 'bg-blue-500 text-white' 
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
              }`}
              onClick={() => setViewMode('preview')}
            >
              Preview
            </button>
            <button 
              className={`px-2 py-1 text-xs rounded ${
                viewMode === 'raw' 
                  ? 'bg-blue-500 text-white' 
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
              }`}
              onClick={() => setViewMode('raw')}
            >
              Raw
            </button>
          </div>
        </div>
      </div>

      {/* Preview Content */}
      <div className="flex-1 overflow-auto">
        {isLoading ? (
          <div className="p-8 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          </div>
        ) : viewMode === 'raw' ? (
          <div className="p-4">
            <pre className="text-sm font-mono bg-gray-50 p-4 rounded border overflow-auto">
              {JSON.stringify(interpreterResult, null, 2)}
            </pre>
          </div>
        ) : (
          <div className="h-full">
            {renderedContent}
          </div>
        )}
      </div>
    </div>
  )
} 