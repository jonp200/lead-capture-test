# Implementation notes

## Decisions

- I prompted the use of the CSV file for the domain structure for a quickstart, since I only got one-hour time allowance for the implementation.
- Only two endpoints I required to build: getting of leads and adding of a new lead. This simplifies the backend's scope.
- I instructed the AI to properly handle the PIIs in the logs for data privacy and security.
- Consequently, I used `log/slog` and write the logs to a file to record helpful info or errors when running a build in a production environment. I encountered an issue when importing a CSV file, where the target file I'm supplying to the command is not getting the expected data as input and I don't see the actual logs.
- Related to the above, I parameterized the target file for the import command, so you won't need to rename files when doing new import tasks.
- I defaulted the environment to production, so when the app runs as a build, you can be sure to get see the logs in the log file (default `logs/leads.log`.)
- I used Lumberjack for the log file rotation, so you won't have problems reading a log file when it grew so much.
- I included the `golangci-lint` configuration file in the prompt to guide the code generation, so there would be minimal fixes later, if ever. The same reason with the prompt about deprecated packages or functions usage.
- I used State constant values for consistency with the input mapping and validation.

## AI suggestions rejected and amended implementations

- The AI previously registered the frontend handler alongside the none-existing route/page handler. I instructed it to move it into its own handler registry to fix that wrong assumption with a none-existing route / page handler registry.
- Minor validation message with the email in the lead form: `already exists` was only displayed. I changed it to `Email already exists` to better reflect what it is pertaining.

## What could have been?

- I would have removed the instruction to review the `README.md` file and let the AI ask me questions as I expected it to do. It implemented way more than what I asked for, resulting to wider code coverage and technical exam code review purpose.
- I would have designed the domain structure myself and feed them to the prompt. The database generated only used a single table, while the other columns are serialized as a text column. Filtering works, but you can get the most benefit if they were created as ordinary columns. Also, database migrations would be more efficient if the other fields were structured as their own column or an entirely separate table.
- Consequently, I would have instructed the use of `GORM` library for the database operations.

## Next steps

- Leads pagination and filtering through the backend.
- Use cent value for processing and storing financial data.
- Backend service health checking as a dependency to the frontend startup.
- Use a modal dialog for the CSV import interface.
- Dismissible action result messages. The message currently stays on the page.
