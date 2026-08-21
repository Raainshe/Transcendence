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
- !TODO: add Database schema + Tables

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
	- !TODO: link to swagger UI

## Modules:
### Major:
- Frontend + Backend Framework 2pts ()
- WebSocket 2pts ()
- Interact with other users 2pts (nmannar)
- Public API 2pts ()
- Standard User Management 2pts ()
- Advanced user roles 2pts ()
- Game (web-based) 2pts ()
- Remote players 2pts ()
- Multiplayer 3+ 2pts ()
- 18 points
### Minor:
- File Upload 1pt ()
- Multi Browser support 1pt ()
- Multi Language support 1pt ()
- Game Stats + Match History 1pt ()
- 2FA 1pt (nmannar)
- Gamification 1pt ()
- 6 points

#### Total: 24 Points

## Individual Contributions:
- jwardeng:
	-
- ksinn:
	-
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
	-
