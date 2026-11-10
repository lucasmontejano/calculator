# Calculator frontend

React and TypeScript interface for the calculator. It sends every operation to the Go API through `/api`.

```bash
npm install
npm run dev        # http://localhost:5173
npm test           # unit and random tests
npm run lint
npm run coverage   # terminal summary + HTML report in ../docs/coverage/frontend
```

The dev server forwards `/api` to `http://localhost:8080`. If the backend runs somewhere else, set `VITE_BACKEND_URL` before starting the dev server.

The root [README](../README.md) covers the full setup, the API and the design decisions.
