describe('Relay App', () => {
  it('should display the welcome message and show the New Project modal', () => {
    cy.visit('http://localhost:5173');
    cy.get('[data-cy=welcome-screen]').should('be.visible');
    cy.contains('New Project').click();
    cy.get('[data-cy=modal-input]').should('be.visible');
    cy.get('[data-cy=modal-create]').should('be.visible');
    cy.get('[data-cy=modal-cancel]').should('be.visible');
  });

  it('should create a new project and navigate to the workspace', () => {
    cy.visit('http://localhost:5173');
    cy.contains('New Project').click();
    cy.get('[data-cy=modal-input]').type('Test Project');
    cy.get('[data-cy=modal-create]').click();
    cy.get('[data-cy=workspace]').should('be.visible');
    cy.get('[data-cy=project-name]').should('contain', 'Test Project');
    cy.get('[data-cy=project-path]').should('contain', '/projects/test-project');
    cy.get('[data-cy=sidebar]').should('be.visible');
    cy.get('[data-cy=live-preview-container]').should('be.visible');
  });
}); 