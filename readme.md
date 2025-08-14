# Lukyasha Football Tracker

Lukyasha Football Tracker is a lightweight football (soccer) events tracker built with Go, Gin, GORM (SQLite), and HTMX. It focuses on simple event management for friendly games with fast, interactive UI updates without heavy front‑end frameworks.

## Features

- Events: create events with name, date, and a link; list and delete events.
- Teams & Players: add teams to an event, add players to teams, quick delete; inputs reset after submit.
- Games: create games between event teams, view game page with scoreboard.
- Goals & Assists: record goal minute and type (normal, penalty, own goal). Optionally link an assist. Players can be picked from any team (useful for mixed/friendly games).
- Timeline: goals and their assist appear as a single row in order of creation; delete goal also deletes linked assist and updates the score.
- Standings: auto‑computed table by event (P, W, D, L, GF, GA, GD, Points) sorted by points, GD, GF, name.
- Leaderboards: top scorers (normal + penalty) and top assistants across the event.
- Live UI with HTMX:
  - Add team updates the game dropdowns instantly (OOB swaps).
  - Delete game re-computes standings and leaderboards without page refresh.
  - Delete team/player removes cards/items inline.
  - Toasts after deletes via HX-Trigger events.
- Mobile friendly: glass navbar, bottom tab bar, larger tap targets, subtle animations.
- Dark/Light theme toggle with persistence.

## Tech Stack

- Go + Gin web framework
- GORM ORM with SQLite (modernc.org/sqlite; no CGO required)
- HTMX for progressive enhancement and partial updates
- Bootstrap 5 for layout and components

## Getting Started

Prerequisites:
- Go 1.21+ (recommended)

Run locally:

```
go run .
```

Then open http://localhost:8080 in your browser.

Notes:
- The app creates `data.db` (SQLite) in the project root on first run.
- AutoMigrate runs at startup; no manual migrations are required.

## Project Structure

- `main.go` – server boot, routes, static files, template loading, DB init
- `models/` – GORM models:
  - `Event`, `Team`, `Player`, `Game`, `GamePlayerStat`
  - `GamePlayerStat` fields include `Type` (goal, penalty, own_goal, assist) and `Minute`
- `handlers/` – HTTP handlers for events, teams, players, games, and stats
- `templates/` – HTML templates (composition via shared partials)
  - `event_detail.html`, `game_detail.html`, `events.html`, etc.
  - Partials: `event_games_list.html`, `event_stats.html`, `team_card.html`, `player_item.html`, `game_goals_list.html`, `team_options.html`
- `static/` – global styles and small client script
  - `app.css` – theme, components, animations
  - `app.js` – theme toggle + toast listener

## Key Interactions (HTMX)

The app uses HTMX to keep pages responsive without full reloads:

- Add Team (on event page):
  - `POST /teams` returns a team card and triggers `team-added` to refresh game selects via `GET /events/:id/team_options` (OOB swap of `<select>` options).
- Add Player (inside team card):
  - `POST /players` returns a new `<li>`; the form resets after submission.
- Add Goal (game page):
  - `POST /games/:id/goals` accepts `team_id`, `player_id`, optional `assist_player_id`, `goal_type`, `minute`. It updates the timeline list and the scoreboard via OOB swap.
- Delete Goal/Assist:
  - Deleting a goal also deletes the linked assist and decrements the score.
- Delete Game:
  - `DELETE /games/:id` removes the game and its stats, triggers `game-removed`, and the event page refreshes its games list, count badge, standings, and leaderboards.
- Delete Event:
  - `DELETE /events/:id` removes the event and all related data in a transaction. From the events list it deletes inline and shows a toast; from the detail page it redirects to `/events`.

## Routing Overview

- `GET /` – Home
- `GET /events` – Events list
- `GET /events/new` – Create event form
- `POST /events` – Create event (HTMX friendly)
- `GET /events/:id` – Event detail (teams, games, standings, leaders)
- `DELETE /events/:id` – Delete event (transactional)
- `POST /teams` – Create team (emits `team-added`)
- `DELETE /teams/:id` – Delete team
- `POST /players` – Create player
- `DELETE /players/:id` – Delete player
- `POST /games` – Create game (via form)
- `GET /games/:id` – Game detail (scoreboard + timeline)
- `POST /games/:id/goals` – Add goal (+optional assist)
- `DELETE /stats/:id` – Delete stat (goal/assist); updates score if needed
- Partials for HTMX:
  - `GET /events/:id/team_options` – OOB refresh for game team selects
  - `GET /events/:id/games_partial` – Games list
  - `GET /events/:id/stats_partial` – Standings + leaderboards

## Data Model Highlights

- A `GamePlayerStat` record captures either a goal (normal/penalty/own_goal) or an assist; goals optionally link the assist via `GoalStatID` so the UI can render them as a single row.
- Scores are persisted in `Game` and kept in sync when adding/removing goals.

## Theming & UX

- Toggle dark/light theme; persisted in `localStorage`.
- Mobile bottom tab bar and glass navbar.
- Toast messages for destructive actions (delete game/event) via `HX-Trigger` and a small app.js listener.
- Subtle animations for added/removed elements.

## Development Tips

- Restart with `go run .` after changing Go code; templates/static changes usually require a hard refresh.
- If CSS/JS appears stale, use Ctrl/Cmd+Shift+R to bypass cache.
- The project uses the modernc SQLite driver, so it works without CGO.

## Roadmap Ideas

- Edit existing timeline entries (minute/type); undo for deletes
- Per-team leaderboards; per-player stats pages
- Import/export event data (JSON/CSV)
