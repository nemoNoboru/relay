import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';

// Test suite for the App component
describe('App Component', () => {
  test('renders welcome message', () => {
    render(<App />);
    const welcomeMessage = screen.getByText(/Welcome to Relay/i);
    expect(welcomeMessage).toBeInTheDocument();
  });

  test('renders New Project and Open Project buttons', () => {
    render(<App />);
    const newProjectButton = screen.getByRole('button', { name: /New Project/i });
    const openProjectButton = screen.getByRole('button', { name: /Open Project/i });
    expect(newProjectButton).toBeInTheDocument();
    expect(openProjectButton).toBeInTheDocument();
  });
});