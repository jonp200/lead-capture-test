Generate the backend and frontend for this project that imports and stores leads from an uploaded CSV file and the ability to add a new lead. Use the sample `leads.csv` for the domain structure.

Provide the API endpoints for the following actions:
- Getting a list of leads
- Adding a new lead

Notes on the leads CSV file:
- A row item can have invalid or missing values
- Invalid row items should be accounted for review and correction
- Observe proper handling for PII in the logs

Notes on code generation:
- Review the `README.md` file and ask no more than seven questions for any clarification or additional input needed
- Align with the `.golangci.yml` linters, so there would be minimal fixes to do after the backend code generation
- Avoid using deprecated packages or functions

Notes for the Go backend:
- Use Gin for the HTTP
- Use SQLite for the database and `modernc.org/sqlite` for pure Go driver

Notes for the Vue frontend:
- Use Vue 3 with TypeScript
- Use Tailwind CSS