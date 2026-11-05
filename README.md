# Lead capture: coding test

Thanks for taking the time to do this. It should take about **an hour**. Use AI coding tools in the creation of the app: we're interested in the decisions you make, and you'll be asked to walk us through the design and AI prompts in the follow-up interview.

## The scenario

Partners send us leads as CSV files. Build a small app that imports and stores leads, displays them, and lets a staff member add a new one.

## What we've provided

```
.
├── leads.csv        # Sample partner file. Some rows are invalid.
├── Makefile         # make dev | make build | make test | make import
├── go.mod           # Go module, no code yet
└── web/             # Vite + Vue 3 + TypeScript + Tailwind CSS app shell
    ├── vite.config.ts           # Dev server proxies /api to http://localhost:8080
    └── src/
        ├── App.vue              # Page shell
        └── components/
            ├── LeadsView.vue    # Empty: display the leads
            └── LeadForm.vue     # Empty: add a lead
```

~~There is no Go code yet, and no TypeScript types or API calls. You write those.~~

**Prerequisites:** Go 1.27+, Node 22.18+ (or 24.12+), make.

## Stack

Please use the stack we use day to day:

- **Backend:** Go with [Gin](https://gin-gonic.com/) for HTTP, and SQLite through the pure-Go `modernc.org/sqlite` driver so the build stays CGO-free.
- **Frontend:** Vue 3 with TypeScript (Composition API, `<script setup>`), styled with Tailwind CSS (v4 is already set up in `web/src/style.css`).

You may add other Go modules or npm packages if you need them.

## The task

1. **Import:** `bin/leads import leads.csv` loads the file into a SQLite database. Some rows are invalid: skip them, report why (with the line number), and import the rest.
2. **API:** `GET /api/leads` returns all leads. `POST /api/leads` saves a new lead and returns it. Use the JSON contract below.
3. **Frontend:** display the leads, and provide a way to add a lead. Show validation errors from the API.
4. **One binary:** `make build` produces a single `bin/leads` that serves both the frontend and the API (`bin/leads serve`). `make build` will fail until you've written the Go side.
5. **Tests** for your validation, and a short `NOTES.md` covering:
   - the decisions you made and why;
   - what you'd do next with more time;
   - one suggestion from your AI assistant that you rejected, and why.

### Rules for a lead

- First name, last name, email, finance amount and state are required.
- Finance amount is in dollars and must be greater than 0.
- State must be a valid Australian state or territory.
- Salary is optional. If given, it must be greater than 0.
- The vehicle is optional. If any vehicle detail is given, the make is required; the model and year stay optional.
- The vehicle year must be between 1990 and next year.
- Emails must be unique.

The same rules apply to the CSV import and to `POST /api/leads`.

## JSON contract

A lead:

```json
{
  "id": 12,
  "first_name": "Zoë",
  "last_name": "Martin",
  "email": "zoe.martin@example.com",
  "finance_amount": 38500.5,
  "salary": null,
  "state": "QLD",
  "vehicle": null,
  "created_at": "2026-11-03T10:15:00+11:00"
}
```

A lead with a partial vehicle:

```json
"vehicle": { "make": "Kia", "model": "EV6", "year": null }
```

- `GET /api/leads` returns `200` with an array. It must be `[]` when there are no leads.
- `POST /api/leads` takes the same shape without `id` and `created_at`, and returns `201` with the saved lead.
- Malformed JSON returns `400`. A validation failure returns `422` with an error per field:

```json
{
  "errors": {
    "finance_amount": "must be greater than 0",
    "vehicle.make": "is required when a model or year is given"
  }
}
```

## Running it

```sh
make dev       # Go API on :8080 and the Vite dev server on :5173 (once you've written main.go)
make build     # bin/leads: one binary containing the frontend and the API
make import    # build, then import leads.csv
make test      # run the Go tests
```

## Stretch goals (optional)

- Filter by state.
- Client-side validation that matches the server rules.
- `409` on a duplicate email.
- Sorting.

## Submitting

Push your work to a repository we can read (or send us a zip), including `NOTES.md`. Please commit as you go rather than in one final commit: we're interested in how you got there.
