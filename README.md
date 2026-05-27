# Chuchote

Application de messagerie temps réel construite avec Go (backend hexagonal) et React TypeScript (frontend), communicant via WebSocket.

---

## Stack technique

| Côté | Technologie |
|------|-------------|
| Backend | Go 1.25, `coder/websocket`, JWT, bcrypt |
| Frontend | React 19, TypeScript, Zustand, Vite, Lucide React |
| Transport | WebSocket (protocole JSON maison) |
| Stockage | In-memory (prêt pour un vrai DB via ports) |

---

## Architecture

Le backend suit l'architecture hexagonale (Ports & Adapters) :

```
Back/
├── domain/          — Modèles métier (User, Room, Message, Connection)
├── application/
│   └── service/     — Logique métier (AuthService, RoomService, MessageService)
├── port/            — Interfaces inbound / outbound
├── adapter/
│   ├── inbound/
│   │   ├── http/    — Handlers HTTP + middleware JWT
│   │   └── ws/      — Handler WebSocket, Client, Parser
│   └── outbound/
│       ├── memory/  — Repositories in-memory
│       └── hub/     — Hub de broadcast par room
└── infrastructure/  — Config, démarrage serveur HTTP
```

Le frontend est organisé autour de stores Zustand et de hooks métier :

```
client/src/
├── store/           — auth.store, room.store, chat.store
├── services/        — websocket.service, auth.service, http.service
├── hooks/           — useAuth, useRooms, useMessages, useTypingIndicator
├── components/
│   ├── layout/      — AuthScreen, AppShell, Sidebar
│   ├── chat/        — ChatWindow, MessageList, MessageBubble, MessageInput
│   └── ui/          — Avatar, Spinner
└── types/           — domain.ts, ws-events.ts, api.ts
```

---

## Fonctionnalités

- **Auth** — Register / Login avec JWT (24h), mot de passe bcrypt
- **Rooms** — Création et liste de rooms, rejoindre en temps réel
- **Messages** — Envoi, édition, suppression, historique au join
- **Optimistic UI** — Message affiché immédiatement avant l'ACK serveur
- **Présence** — Indicateur de frappe temps réel (typing indicator)
- **Reconnexion** — Le client WebSocket se reconnecte automatiquement

---

## Protocole WebSocket

Chaque frame est un JSON de la forme :

```json
{
  "type": "message.send",
  "requestId": "uuid-or-null",
  "roomId": "uuid-or-null",
  "payload": {}
}
```

| Direction | Événements |
|-----------|-----------|
| Client → Serveur | `auth.connect`, `room.join`, `room.leave`, `message.send`, `message.edit`, `message.delete`, `typing.start`, `typing.stop` |
| Serveur → Client | `auth.connected`, `auth.error`, `room.joined`, `message.ack`, `message.new`, `message.edited`, `message.deleted`, `typing.indicator` |

---

## Lancer le projet

### Backend

```bash
cd Back
JWT_SECRET=mon_secret go run ./main.go
# Démarre sur :8080 par défaut
```

Variables d'environnement :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8080` | Port d'écoute |
| `JWT_SECRET` | `change-me-in-production` | Clé de signature JWT |
| `DATABASE_DSN` | `` | DSN pour future base de données |

### Frontend

```bash
cd client
npm install
npm run dev
# Démarre sur :5173
```

---

## Routes HTTP

```
POST  /auth/register   — Créer un compte
POST  /auth/login      — Se connecter (retourne un JWT)
GET   /rooms           — Lister les rooms (JWT requis)
POST  /rooms           — Créer une room (JWT requis)
GET   /rooms/{id}      — Détails d'une room (JWT requis)
GET   /ws              — Upgrade WebSocket
```

---

## Documentation

`WEBSOCKET_COURS.md` — cours complet (en français) sur les WebSockets : handshake HTTP, format des frames, pattern Hub, UI optimiste, comparaison avec SSE/gRPC, etc. Basé sur le code de ce projet.

---

## Licence

MIT
