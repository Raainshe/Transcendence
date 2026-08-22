*This project has been created as part of the 42 curriculum by jwardeng, ksinn, nmannar, rmakoni*

## Description
This project is a modern, web-based implementation of the classic arcade game Tetris.

## Instructions

```bash
make up
```

Open **https://localhost**. The browser will warn about Caddy’s self-signed certificate; proceed anyway (Advanced → Continue / Accept the Risk).

All client-facing backend traffic (REST, public API, swagger, uploads, WebSockets) is HTTPS/WSS on port 443. Port 80 only redirects to HTTPS. Ports 8080 and 5173 are not published.

```bash
make down    # stop containers
make logs    # tail logs
make help    # other targets
```

## Resources
!TODO: add you resources
- Go: [Docs](https://go.dev/doc/)
- AI usage: Go tests


## Team Information: !TODO: add brief descriptions
jwardeng: Developer
ksinn: Product Owner + Developer
nmannar: Developer
rmakoni: Project Manager + Technical Lead + Developer

## Project Management:
- How the team organized the work:
	- talked about what still needs to be done -> assigned tasks
- Tools used for project management: GitHub, Pull Requests
- Communication channels used: Discord

## Technical Stack:
- Frontend: Vue.js + Vite + Pinia
- Backend: Go + Chi
- Database: Postgres17, because it's the most popular
- Other: Goose + Docker + JWT + i18n + SendGrid
- Justifications: !TODO: add

## Database Schema:
- [migrations folder](./backend/migrations/)

## Features List:
- User:
	- Profile Picture
	- select language
	- edit username
	- enable 2fa
	- create / revoke API Key
	- delete Account
- Friends:
	- add / remove friends
	- accept / deny friend request
	- block friends
	- chat with friends
- Footer:
	- list all players
	- view player profiles
	- view privacy policy
	- view terms
	- view developers
- Game:
	- create new game
	- select game mode (4 modes)
	- play multiplayer
		- create / join lobby (up to 4 players)
	- enable / disable music
- Stats:
	- Leaderboard
	- player statistics
	- match history
- Gamification:
	- progression bar (level ups)
	- badges / achievements
- Public API:
	- [View API in Swagger UI](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/Raainshe/Transcendence/refs/heads/main/openapi.yaml)

## Modules:
### Major:
- Frontend + Backend Framework 2pts (Frontend: rmakoni, Backend: ksinn)
- WebSocket 2pts ()
- Interact with other users 2pts (nmannar, ksinn)
- Public API 2pts ()
- Standard User Management 2pts (ksinn)
- Advanced user roles 2pts ()
- Game (web-based) 2pts ()
- Remote players 2pts ()
- Multiplayer 3+ 2pts ()
- 18 points
### Minor:
- File Upload 1pt (ksinn)
- Multi Browser support 1pt ()
- Multi Language support 1pt ()
- Game Stats + Match History 1pt (ksinn)
- 2FA 1pt (nmannar)
- Gamification 1pt ()
- 6 points

#### Total: 24 Points

## Individual Contributions:
- jwardeng:
	- Implemented the public API module, including the functionality for a user to create an API-Key: the key is hashed with SHA-256 before being stored in the DB. Added a 2-phase rate-limiter: one IP-based, which kicks in when unauthenticated traffic hits the endpoints, and one API-Key-based, used to prevent abuse and enforce per-key limits for API-Key holders.The public endpoints, mirror already existing routes, so I added a public DTO for each resource to control what's exposed. I used swagger to document the public API and enable a quick overview & try-out section and added a Lobby-Name option for lobbies, so the public PUT endpoint would have a more meaningful field to expose.
	- Added the gamification feature, which includes persistent storage and visual feedback for User Achievements, Badges, XP and Leveling. Achievements are based on milestones like score and total-points thresholds, level reached, games played/won, and actions such as adding a friend or changing your avatar. XP increases based on Achievements unlocked, Wins and Games Played. Designed the badges to give visual feedback for achievements.
	- Challenges: Since I joined the team a little bit later on my main challenge was to learn the code-base and get used to the existing tech-stack. I had to learn Go on the go and get accustomed to the frontend stack. It was a really good experience tho thanks to the good architecture of my teammates and the Go's C-like syntax.

- ksinn:
	- Implemented the Go Backend:
		- Go module, base structs, first SQL migrations
		- Chi router + route registration, JWT auth middleware
		- auth/user/game handler-service
	- File uploads:
		- Avatar upload endpoint + underlzing file storage implementation
	- OpenAPI specs
	- Friend system:
		- Model, DB Layer, service functions and routes for the friends system
	- Backend unit test suite
	- Cleanup / best practics:
		- unified error responses, stopped silentlz swallowing errors
		- tighten password policy, request body size capping, made db health checks log and cintinue instead of stopping hard
	- Integration test suite
	- User online status
		- added last seen claim to JWTs
- nmannar:
	- Implemented optional email 2FA. On the backend, added an AuthService extension
      that uses bcrypt-hashed 6-digit codes with a 10-minute expiry,
      sends them through an SMTP mailer, and uses a verification step for login.
      On the frontend, built the two-step login, the enable/disable 2FA controls in the profile page.
      Integrated it with SendGrid for delivery.
    - Built persistent 1-on-1 messaging between friends. On the backend added a chat service
      that enforces friends-only access, validates and length-caps messages,
      and sends them over the existing WebSocket hub. History and unread counts are stored.
      On the frontend built the ChatView (conversation sidebar + message thread).
    - Challenges: When implementing Chat, there was a tough bug where clicking "Send" did nothing.
      I traced it through the browser's inspector and found the cause: a Go nil slice caused .map()
      in the frontend to throw before the socket connected. Had minimal experience with Go, Typescript,
      Vue and SQL before this project.
- rmakoni
	- Built most of the Vue frontend: routing, layout, menus, play view, HUD,
	  and the lobby/match UI on top of Pinia and the existing auth/i18n stack.
	- Designed the Tetris engine as framework-agnostic TypeScript, following the
	  Tetris Guideline: 7-bag generation, SRS wall-kicks, DAS/ARR input,
	  extended lock-down, gravity/levels, Hold, ghost piece, Guideline scoring,
	  T-Spins, Back-to-Back, and Marathon / Sprint / Ultra / multiplayer variants.
	- Implemented remote multiplayer over WebSockets. Clients stream player.state
	  (board, score, lines) into a match room; opponents render live on the play
	  view, with reconnect/backoff and elimination / match-ended handling.
	- On the backend, built lobby flow: create and join by invite code (2–4 players),
	  ready/start gates, host-only start, and handing the lobby off onto a match
	  WebSocket room when the game begins.
	- I struggled the most with developing the engine as I had various bugs which had to do with the logic I implemented and the game not behaving correctly. I also struggled with implementing the Websockets for the game and wanted a more complex UI but required a rewrite.
