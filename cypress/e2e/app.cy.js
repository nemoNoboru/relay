describe('Relay App', () => {
  it('should display the welcome message and handle New Project button', () => {
    // Visit the app
    cy.visit('http://localhost:5173');

    // Check for the welcome message
    cy.contains('Welcome to Relay').should('be.visible');

    // Click the New Project button
    cy.contains('New Project').click();

    // Verify the prompt is shown
    cy.on('window:prompt', (text) => {
      expect(text).to.contains('Enter project name:');
    });
  });
}); 