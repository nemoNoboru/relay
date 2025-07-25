import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import LivePreview from '../LivePreview';

// Mock the Lua runtime result structure
const mockLuaResult = {
  components: [
    {
      type: 'component',
      name: 'input',
      props: {
        name: 'test-input',
        placeholder: 'Enter text'
      }
    },
    {
      type: 'component',
      name: 'button',
      props: {
        text: 'Click me',
        onClick: 'test-click'
      }
    }
  ]
};

describe('LivePreview Component', () => {
  const defaultProps = {
    interpreterResult: null,
    eventHandlers: new Map(),
    onEventTriggered: jest.fn(),
    formValues: {},
    onFormValueUpdate: jest.fn()
  };

  test('renders empty state when no content', () => {
    render(<LivePreview {...defaultProps} />);
    expect(screen.getByText('No Content')).toBeInTheDocument();
    expect(screen.getByText('Start writing Lua code to see the preview')).toBeInTheDocument();
  });

  test('renders error state when error is provided', () => {
    const errorProps = {
      ...defaultProps,
      error: 'Test error message'
    };
    render(<LivePreview {...errorProps} />);
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });

  test('renders loading state when isLoading is true', () => {
    const loadingProps = {
      ...defaultProps,
      isLoading: true
    };
    render(<LivePreview {...loadingProps} />);
    // The loading spinner should be present
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  test('renders components from Lua result', () => {
    const propsWithResult = {
      ...defaultProps,
      interpreterResult: mockLuaResult
    };
    render(<LivePreview {...propsWithResult} />);
    
    // Should render the input and button from the Lua result
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  test('switches between preview and raw view modes', () => {
    const propsWithResult = {
      ...defaultProps,
      interpreterResult: mockLuaResult
    };
    render(<LivePreview {...propsWithResult} />);
    
    // Initially in preview mode
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
    
    // Switch to raw mode
    const rawButton = screen.getByText('Raw');
    fireEvent.click(rawButton);
    
    // Should show JSON representation
    expect(screen.getByText(/"name": "test-input"/)).toBeInTheDocument();
  });

  test('handles input value changes', () => {
    const mockOnFormValueUpdate = jest.fn();
    const propsWithResult = {
      ...defaultProps,
      interpreterResult: mockLuaResult,
      onFormValueUpdate: mockOnFormValueUpdate
    };
    render(<LivePreview {...propsWithResult} />);
    
    const input = screen.getByPlaceholderText('Enter text');
    fireEvent.change(input, { target: { value: 'test value' } });
    
    expect(mockOnFormValueUpdate).toHaveBeenCalledWith('test-input', 'test value');
  });

  test('handles button clicks', () => {
    const mockOnEventTriggered = jest.fn();
    const propsWithResult = {
      ...defaultProps,
      interpreterResult: mockLuaResult,
      onEventTriggered: mockOnEventTriggered
    };
    render(<LivePreview {...propsWithResult} />);
    
    const button = screen.getByText('Click me');
    fireEvent.click(button);
    
    expect(mockOnEventTriggered).toHaveBeenCalledWith('test-click', {});
  });
}); 