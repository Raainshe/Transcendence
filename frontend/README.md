# Frontend

Vue 3 SPA for Transcendence: menus, profiles, friends, chat, and a Tetris Guideline engine with live 2–4 player matches.

The evaluated stack is the repo root (`make up` → **https://localhost:8443**). This README covers working on the client itself.

## Stack

- Vue 3, Vite, TypeScript
- Vue Router, Pinia, vue-i18n (en / de / fr)
- Tailwind CSS v4 (utilities only; Preflight off so the BEM theme stays intact)
- Vitest, ESLint, oxlint, Prettier

Requires Node `^20.19.0 || >=22.12.0`.

## Scripts

```bash
npm install
npm run dev          # Vite on http://localhost:5173 (proxies /api and /uploads)
npm run build        # type-check + production bundle
npm run preview      # serve the production build
npm run test         # Vitest (watch)
npm run test:run     # Vitest once
npm run lint
npm run format
```

For host-side `npm run dev`, the API should already be running. Vite proxies `/api` and `/uploads` to `API_PROXY_TARGET` (default `http://localhost:8080`). Optional variables are in [`.env.example`](./.env.example).

Behind Caddy, `HMR_CLIENT_PORT` is set in Compose so Hot Module Replacement uses WSS. Do not point `VITE_API_BASE_URL` at a plain `http://` origin for the evaluated HTTPS stack.

## Layout

| Path | Role |
| --- | --- |
| `src/views/` | Route pages (home, play, lobby, profile, friends, chat, …) |
| `src/components/` | Layout, menu, game HUD, auth, profile |
| `src/game/` | Framework-agnostic Tetris engine, input, scoring, render |
| `src/stores/` | Pinia: auth, lobby, match, session, settings |
| `src/composables/` | WebSocket clients (match, lobby, chat) |
| `src/api/` | REST helpers |
| `src/assets/styles/` | Design tokens, reset, BEM stylesheets |
| `src/locales/` | i18n JSON |
| `src/router/` | Vue Router |

`@` aliases to `src/` (see `vite.config.ts`).

## IDE

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (disable Vetur). Vue DevTools in Chromium or Firefox is optional but useful.
