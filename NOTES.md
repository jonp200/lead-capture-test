# Implementation notes

## Decisions

- Gin handles the API; SQLite uses the CGO-free `modernc.org/sqlite` driver. A single embedded Vue build is served from the same binary and origin. No frontend runtime server is needed after building.
- One normalization/validation path is shared by CSV and JSON requests. Names are trimmed, email is lowercased, and state abbreviations/full names are canonicalized. CSV money accepts a leading dollar sign and commas, including the supplied `$45,000.00`. Nonfinite numbers are rejected. Blank salary, vehicle and vehicle year preserve JSON `null`.
- SQLite enforces email uniqueness with `COLLATE NOCASE`, so duplicate detection remains correct for concurrent requests. Duplicate API submissions return `409` with an email field error. Leads are stored as JSON with a separate indexed email and authoritative SQLite ID; this keeps the small domain simple. List responses use insertion order, newest first.
- Each valid CSV row commits independently. Invalid rows do not block subsequent records. Reports contain physical source line numbers and safe field errors. The original CSV remains the correction source; rejected personal data is not copied into logs or another database. Upload errors include import progress because earlier accepted rows remain committed. Re-uploading an entire file safely skips existing emails.
- CSV upload uses `POST /api/leads/import` with the raw file as a `text/csv` body, limited to 10 MiB. JSON bodies are limited to 64 KiB and must contain one object. Unknown JSON fields are rejected. Read-only metadata on input is ignored and regenerated.
- The UI includes loading, empty, network error, success and import review states, native dialog keyboard handling, labeled controls, field errors, name/email search, state filtering and name sorting. Server validation is authoritative; native browser constraints catch common input errors.
- Request bodies, email addresses, file names, raw CSV records and database errors are not logged. API responses use `Cache-Control: no-store`. The app binds to loopback by default. This is a local staff prototype, without authentication or encrypted database storage.
- The existing strict lint rules remain enabled. The only path exclusion added is third-party Go source inside `web/node_modules`; CLI file access has a specific suppression because accepting a user-selected path is its intended behavior.

## AI suggestion rejected

Using `gin.Default()` would provide convenient request logs and panic recovery. I rejected that option during AI-assisted review: recovery diagnostics can include request headers, which conflicts with the PII requirement. The application instead uses Gin without request logging and recovery with a fixed public error message. This decision was made by the coding agent; it is not a claim that a human reviewer rejected a suggestion.

## Verification

Validation tests cover required fields, malformed emails, zero/negative/nonfinite amounts, all states, optional salary and vehicles, dependent make validation, and year boundaries. Integration tests cover JSON errors, duplicate casing, generated IDs, empty arrays, persistence across database reopen, frontend serving, CSV uploads, malformed rows, headers, Unicode, and multiline physical line numbers. The supplied CSV yields 36 accepted rows and four rejections: lines 25 (state), 29 (email), 33 (vehicle make), and 37 (duplicate email).

The Vue production build/type check, frontend ESLint, Go tests and configured Go lint checks pass (zero lint issues). A CGO-disabled Windows executable successfully imported the sample. Browser checks verified the embedded UI, Unicode search, duplicate email field errors, adding a lead, state filtering and CSV re-upload reporting. Test records were confined to the ignored `bin/smoke.db`. GNU make is unavailable on this host, so its build steps were run directly.

## Next steps

- Add staff authentication, authorization, deployment-specific TLS, restricted database permissions, encrypted storage/backups and a retention policy before handling real customer data in a shared deployment.
- Add a durable import batch/review workflow and an explicit correction endpoint if staff need to edit rejected rows within the app. Currently staff correct the original file and re-upload it, or use the add form.
- Move to integer cents/decimal validation before doing financial calculations; current amounts follow the supplied floating-point JSON contract and are displayed as AUD.
- Add schema migrations, pagination and server-side filtering for larger datasets, structured PII-free operational metrics, plus automated browser interaction tests.
- An unterminated quoted CSV field can consume following physical lines under standard CSV parsing. Reports identify the start of that malformed record; correct the quoting and retry the source file.

## Running on this Windows machine

GNU make is not installed here. The equivalent build commands are:

```powershell
Push-Location web
npm ci # needed only on a fresh checkout
npm run build
Pop-Location
$env:CGO_ENABLED = '0'
go build -o bin/leads.exe .
./bin/leads.exe import leads.csv
./bin/leads.exe serve
```

Open `http://127.0.0.1:8080`. Set `LEADS_DB` to choose the SQLite path (default `leads.db`) and `LEADS_ADDR` to change the listen address. Run `go test ./...` after building the frontend, `golangci-lint run ./...`, and `npm run lint` in `web`. On systems with GNU make, the original `make build`, `make import`, `make dev`, and `make test` targets work unchanged.

