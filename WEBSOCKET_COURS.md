# WebSocket — Cours complet

> **Contexte** : ce document s'appuie sur le projet **Chuchote**, une application de messagerie temps réel construite avec Go (backend) et React TypeScript (frontend). Chaque concept est illustré par le vrai code du projet.

---

## Sommaire

1. [Le problème que WebSocket résout](#1-le-problème-que-websocket-résout)
2. [Ce qu'est WebSocket exactement](#2-ce-quest-websocket-exactement)
3. [L'upgrade HTTP → WebSocket (le handshake)](#3-lupgrade-http--websocket-le-handshake)
4. [La structure d'une frame WebSocket](#4-la-structure-dune-frame-websocket)
5. [Comment on envoie un message — du clic au stockage](#5-comment-on-envoie-un-message--du-clic-au-stockage)
6. [Comment les messages sont stockés](#6-comment-les-messages-sont-stockés)
7. [Concevoir un protocole au-dessus de WebSocket](#7-concevoir-un-protocole-au-dessus-de-websocket)
8. [Le pattern UI optimiste](#8-le-pattern-ui-optimiste)
9. [Gestion des connexions multiples — le Hub](#9-gestion-des-connexions-multiples--le-hub)
10. [Cycle de vie complet d'une connexion](#10-cycle-de-vie-complet-dune-connexion)
11. [WebSocket vs les alternatives](#11-websocket-vs-les-alternatives)
12. [Ce qu'il faut retenir pour l'oral](#12-ce-quil-faut-retenir-pour-loral)

---

## 1. Le problème que WebSocket résout

### HTTP classique est unidirectionnel

Le protocole HTTP fonctionne sur un modèle **requête / réponse** :

```
Client ──── GET /messages ────────► Serveur
Client ◄─── 200 OK [données] ─────── Serveur
(connexion fermée)

Client ──── GET /messages ────────► Serveur   ← même requête, 1 seconde après
Client ◄─── 200 OK [données] ─────── Serveur
(connexion fermée)
```

**Règle absolue de HTTP** : c'est toujours le client qui initie. Le serveur ne peut jamais envoyer un message de sa propre initiative. Il ne peut que répondre.

### Le polling — la mauvaise solution historique

Avant WebSocket, pour simuler du temps réel on utilisait le **polling** : le client envoie une requête toutes les X secondes pour demander "y a-t-il du nouveau ?"

```
Client ──── GET /messages?since=12:00:00 ────► Serveur   (12:00:01)
Client ◄─── 200 OK []  ──────────────────────── Serveur   (rien de nouveau)

Client ──── GET /messages?since=12:00:00 ────► Serveur   (12:00:02)
Client ◄─── 200 OK []  ──────────────────────── Serveur   (rien de nouveau)

Client ──── GET /messages?since=12:00:00 ────► Serveur   (12:00:03)
Client ◄─── 200 OK [{ "content": "Salut !" }]  Serveur   ← enfin !
```

**Problèmes du polling :**
- **Latence** : si Alice envoie un message à 12:00:02.5, Bob ne le verra qu'à sa prochaine requête (12:00:03). Le délai moyen est de la moitié de l'intervalle.
- **Charge serveur** : 100 utilisateurs × 1 requête/seconde = 100 requêtes/seconde même quand personne ne parle.
- **Bande passante** : chaque requête HTTP embarque des headers (cookies, User-Agent, Authorization…) qui peuvent peser 500 octets — pour souvent recevoir une réponse vide.

### Long polling — une amélioration partielle

Le long polling garde la connexion ouverte jusqu'à ce qu'un message arrive (ou un timeout) :

```
Client ──── GET /messages ────────────────────► Serveur
                                                 (attend...)
                                                 (attend...)
                                                 (reçoit un événement)
Client ◄─── 200 OK [{ "content": "Salut !" }] ── Serveur
(connexion fermée, le client en ouvre une nouvelle immédiatement)
```

Meilleur, mais toujours des problèmes : reconnexion permanente, overhead HTTP à chaque cycle, difficile à scaler.

### WebSocket — la bonne solution

WebSocket crée une connexion **bidirectionnelle persistante** :

```
Client ══════════════════════════════════════ Serveur
              (connexion ouverte en permanence)

Client ──── "Salut Bob !" ───────────────────► Serveur
Client ◄─── "Reçu 5/5 !"  ─────────────────── Serveur   ← le serveur POUSSE
Client ◄─── "Alice est en train d'écrire…" ─── Serveur   ← sans que le client ait demandé
```

Un seul établissement de connexion. Ensuite les deux parties envoient librement, à tout moment, dans les deux sens.

---

## 2. Ce qu'est WebSocket exactement

### Définition technique

WebSocket est un **protocole de communication** standardisé par le RFC 6455 (2011). Il fonctionne **par-dessus TCP**, comme HTTP, mais avec un protocole de couche applicative différent après le handshake initial.

```
┌─────────────────────────────────────────┐
│           Application (JSON, etc.)       │  ← notre protocole métier
├─────────────────────────────────────────┤
│              WebSocket                   │  ← framing, ping/pong, close
├─────────────────────────────────────────┤
│                 TCP                      │  ← transport fiable
├─────────────────────────────────────────┤
│                  IP                      │  ← routage réseau
└─────────────────────────────────────────┘
```

### Caractéristiques fondamentales

| Propriété | Valeur |
|---|---|
| **Bidirectionnel** | Oui — client ET serveur envoient librement |
| **Full-duplex** | Oui — les deux peuvent envoyer simultanément |
| **Persistant** | Oui — une seule connexion TCP maintenue ouverte |
| **Bas overhead** | Oui — headers de 2 à 10 octets par frame (vs ~500 pour HTTP) |
| **Texte ou binaire** | Les deux — on choisit par frame |
| **Port** | 80 (ws://) ou 443 (wss://) — traversent les firewalls |

### ws:// vs wss://

```
ws://localhost:8080/ws    ← non chiffré, développement local seulement
wss://api.monapp.com/ws   ← TLS (chiffré), obligatoire en production
```

`wss://` est WebSocket sur TLS, exactement comme `https://` est HTTP sur TLS. En production, **toujours** utiliser `wss://`.

---

## 3. L'upgrade HTTP → WebSocket (le handshake)

C'est le mécanisme le plus important à comprendre. WebSocket démarre TOUJOURS par une requête HTTP. Après, HTTP n'est plus utilisé.

### Étape 1 — La requête d'upgrade (client → serveur)

Le navigateur envoie une requête HTTP GET spéciale :

```http
GET /ws HTTP/1.1
Host: localhost:8080
Connection: Upgrade
Upgrade: websocket
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
Origin: http://localhost:5173
```

Détail des headers importants :

- `Connection: Upgrade` — "je veux changer de protocole sur cette connexion TCP"
- `Upgrade: websocket` — "le nouveau protocole que je veux, c'est WebSocket"
- `Sec-WebSocket-Key` — une clé aléatoire encodée en base64, générée par le navigateur. Elle sert à prouver que le serveur a bien reçu la demande (et non pas qu'une réponse HTTP en cache a été recyclée).
- `Sec-WebSocket-Version: 13` — version du protocole (13 est la seule version standard actuelle)

### Étape 2 — La réponse d'acceptation (serveur → client)

Si le serveur accepte, il répond avec le code **101 Switching Protocols** (jamais 200 !) :

```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

Le `Sec-WebSocket-Accept` est calculé ainsi :
```
base64( SHA1( Sec-WebSocket-Key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11" ) )
```
La chaîne `"258EAFA5..."` est une constante magique définie par le RFC. Ce mécanisme prouve que le serveur comprend le protocole WebSocket et a intentionnellement répondu à cette demande.

### Étape 3 — La connexion est ouverte

Après le `101`, la **connexion TCP reste ouverte** mais HTTP est abandonné. La même socket physique est maintenant utilisée par le protocole WebSocket. Le navigateur et le serveur savent tous les deux que les prochains octets qui arrivent seront des **frames WebSocket**, plus des requêtes HTTP.

### Dans notre code

```go
// Back/adapter/inbound/ws/handler.go

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Cette ligne fait tout le handshake décrit ci-dessus automatiquement.
    // Elle lit la requête HTTP, vérifie les headers Upgrade,
    // calcule le Sec-WebSocket-Accept et renvoie le 101.
    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        InsecureSkipVerify: true,
    })
    if err != nil {
        return // handshake refusé
    }

    // À partir d'ici, `conn` est une connexion WebSocket active.
    // HTTP n'existe plus sur cette socket.
    c := newClient(conn, h.hub, h.messages, h.rooms, h.auth)
    c.run(r.Context())
}
```

```typescript
// client/src/services/websocket.service.ts

connect(url: string): void {
    // new WebSocket() déclenche le handshake HTTP → WebSocket automatiquement.
    // Le navigateur envoie les headers Upgrade, attend le 101, puis passe en mode WS.
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
        // onopen se déclenche APRÈS le 101, quand la connexion WS est établie.
        // C'est ici qu'on peut envoyer le premier message.
        const pending = this.queue.splice(0);
        for (const event of pending) {
            this.ws!.send(JSON.stringify(event));
        }
    };
}
```

**Point important** : entre `new WebSocket(url)` et `onopen`, la connexion est en état `CONNECTING`. Si tu appelles `ws.send()` pendant cet état, le message est **silencieusement ignoré**. C'est pourquoi on a une file d'attente (`this.queue`) qui stocke les messages envoyés trop tôt et les expédie dès que `onopen` se déclenche.

---

## 4. La structure d'une frame WebSocket

Une fois la connexion établie, les données circulent sous forme de **frames** — des petits paquets binaires avec un en-tête compact.

### Anatomie d'une frame

> **Comment lire ce diagramme ?**
> Le sens de lecture est **horizontal, de gauche à droite**, bit par bit.
> - La **ligne du haut** (0 · 1 · 2 · 3) = numéro de l'**octet**
> - La **ligne en dessous** (0 1 2 3 4 5 6 7 | 0 1 2 3…) = numéro du **bit** dans l'octet
> - Chaque colonne représente **1 bit**
> - Si un nom de champ est écrit **verticalement** (ex: F/I/N sur 3 lignes), c'est uniquement parce que le mot ne tient pas en largeur dans 1 seul bit — ça reste **un seul champ d'1 bit**

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (si payload len == 126/127) |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+-------------------------------+
|     Extended payload length continued, si payload len == 127  |
+---------------------------------------------------------------+
|                    Masking-key (si MASK == 1)                  |
+---------------------------------------------------------------+
|                    Payload Data                                |
+---------------------------------------------------------------+
```

**Légende des champs :**

| Champ | Taille | Rôle |
|---|---|---|
| **FIN** | 1 bit | `1` = c'est la dernière (ou seule) frame du message. `0` = il y a d'autres fragments qui suivent. |
| **RSV1/2/3** | 1 bit chacun | Réservés pour des extensions futures (compression, etc.). Toujours `0` dans un WS basique. |
| **opcode** | 4 bits | Type de la frame : `0x1` = texte, `0x2` = binaire, `0x8` = fermeture, `0x9` = ping, `0xA` = pong. |
| **MASK** | 1 bit | `1` = le payload est masqué. **Toujours `1` côté client → serveur** (obligatoire RFC 6455). `0` côté serveur → client. |
| **Payload len** | 7 bits | Longueur du payload. Si `< 126` : c'est la vraie longueur. Si `= 126` : lire 2 octets de plus. Si `= 127` : lire 8 octets de plus. |
| **Extended payload length** | 16 ou 64 bits | Présent uniquement si `Payload len == 126` (16 bits) ou `== 127` (64 bits). Donne la vraie longueur pour les grands messages. |
| **Masking-key** | 32 bits (4 octets) | Clé aléatoire utilisée pour masquer le payload. Présente uniquement si `MASK == 1`. |
| **Payload Data** | variable | Le contenu réel du message, XOR-é avec la masking-key octet par octet si `MASK == 1`. |

> **Pourquoi masquer côté client ?** Pour empêcher les proxies et caches HTTP intermédiaires de mal interpréter les données WebSocket comme du HTTP. C'est une protection réseau, pas une protection contre l'espionnage.

En pratique, pour un message texte court (< 126 octets) :

```
Octet 1 : FIN=1 + opcode=0x1 (texte)     → 0x81
Octet 2 : MASK=1 + longueur (ex: 45)     → 0xAD  (client→serveur)
Octets 3-6 : Masking key (4 octets aléatoires)
Octets 7-N : Payload masqué (XOR avec la masking key)
```

**Overhead total : 6 octets** pour un message client→serveur court. Compare ça aux 400-800 octets de headers d'une requête HTTP classique.

### Les opcodes

| Opcode | Valeur | Signification |
|---|---|---|
| Continuation | `0x0` | Suite d'une frame fragmentée |
| Text | `0x1` | Payload texte (UTF-8) |
| Binary | `0x2` | Payload binaire |
| Close | `0x8` | Fermeture de la connexion |
| Ping | `0x9` | Test de vivacité |
| Pong | `0xA` | Réponse au ping |

### Pourquoi le masquage ?

Les messages **client → serveur** sont obligatoirement masqués (XOR avec une clé de 4 octets). Les messages **serveur → client** ne sont PAS masqués.

La raison est sécuritaire : sans masquage, un script malveillant dans un navigateur pourrait envoyer des frames qui ressemblent à des requêtes HTTP vers des proxies intermédiaires et les tromper (**cache poisoning**). Le masquage rend les frames WS impossible à confondre avec du HTTP.

### Ce qu'on utilise réellement dans le code

On ne touche jamais aux octets bruts. La librairie `coder/websocket` abstrait tout :

```go
// Lecture d'un message
var raw json.RawMessage
wsjson.Read(ctx, conn, &raw)  // démasque, ré-assemble les fragments, vérifie l'opcode

// Écriture d'un message
conn.Write(ctx, websocket.MessageText, data)  // frappe, encode en UTF-8, envoie
```

```typescript
// Côté navigateur, l'API WebSocket standard abstrait aussi les frames
this.ws.send(JSON.stringify(event))  // le navigateur crée la frame avec masquage
this.ws.onmessage = (e) => {
    const frame = JSON.parse(e.data)  // `e.data` est déjà le payload démasqué
}
```

---

## 5. Comment on envoie un message — du clic au stockage

Voici le chemin complet d'un message dans Chuchote, de la frappe clavier jusqu'à ce que Bob le voit sur son écran.

### Vue d'ensemble

```
[Alice frappe "Salut Bob !" + Entrée]
         │
         ▼
① MessageInput.tsx
  onChange → met à jour le state local (value)
  onKeyDown Enter → appelle handleSubmit()
  handleSubmit() → appelle send(content) depuis useMessages
         │
         ▼
② useMessages.ts (hook React)
  Crée un message OPTIMISTE (pending: true, id temporaire)
  addMessage(roomId, optimistic) → Zustand store mis à jour
  → UI affiche immédiatement le message en gris
  wsService.send({ type: "message.send", clientTempId, content })
         │
         ▼ (frame WebSocket, JSON encodé, ~150 octets)
         │
         ▼
③ Back/adapter/inbound/ws/client.go (goroutine de lecture)
  wsjson.Read(ctx, conn, &raw) → reçoit la frame
  parseFrame(raw) → décode en WSFrame{ type, roomId, payload }
  switch frame.Type → case "message.send"
  json.Unmarshal → MessageSendPayload{ content, clientTempId }
  messageService.SendMessage(ctx, cmd)
         │
         ▼
④ Back/application/service/message_service.go
  model.NewMessage() → valide (non vide, ≤ 4000 chars) → crée Message avec UUID serveur
  messageRepo.Save(ctx, msg) → STOCKAGE
  broadcastNewMessage(view) → hub.BroadcastToRoomExcept()
         │
         ├──────────────────────────────────────────────────────────────────────────┐
         ▼                                                                          ▼
⑤a message.ack → canal send d'Alice                               ⑤b message.new → canal send de Bob
  { messageId: "uuid-serveur", clientTempId: "tmp-001" }            { id, content, authorName, createdAt }
         │                                                                          │
         ▼                                                                          ▼
⑥a writePump d'Alice (goroutine d'écriture)                       ⑥b writePump de Bob
  conn.Write(ctx, websocket.MessageText, data)                      conn.Write(ctx, websocket.MessageText, data)
         │                                                                          │
         ▼ (frame WebSocket)                                                        ▼ (frame WebSocket)
         │                                                                          │
         ▼                                                                          ▼
⑦a websocket.service.ts (onmessage)                               ⑦b websocket.service.ts (onmessage)
  case "message.ack":                                               case "message.new":
  confirmMessage(roomId, clientTempId, confirmedMsg)                addMessage(roomId, msg)
  → remplace l'optimiste par le msg confirmé                        → ajoute le message de Bob
  → pending: false, id réel                                        → Bob voit "Salut Bob !" apparaître
         │                                                                          │
         ▼                                                                          ▼
  MessageList ré-affiche sans le gris               MessageList ré-affiche avec le message d'Alice
```

### Le code de chaque étape

**① L'input — `client/src/components/chat/MessageInput.tsx`**
```tsx
const handleSubmit = () => {
    const trimmed = value.trim();
    if (!trimmed) return;           // cas limite : message vide ignoré
    send(trimmed);                  // délègue à useMessages
    setValue('');                   // vide l'input
    textareaRef.current?.focus();   // remet le focus pour la prochaine frappe
};
```

**② Le hook — `client/src/hooks/useMessages.ts`**
```typescript
const send = (content: string) => {
    const clientTempId = crypto.randomUUID();  // id temporaire côté client

    // Message optimiste — affiché AVANT la confirmation serveur
    const optimistic: Message = {
        id: clientTempId,       // id temporaire (sera remplacé par l'uuid serveur)
        content,
        authorId: userId!,
        authorName: username!,
        createdAt: new Date().toISOString(),
        pending: true,          // flag visuel → bulle grisée
    };
    addMessage(roomId, optimistic);   // → Zustand → React re-render → UI immédiate

    wsService.send({
        type: 'message.send',
        requestId: crypto.randomUUID(),
        roomId,
        payload: { content, clientTempId },
    });
};
```

**③ La lecture WebSocket — `Back/adapter/inbound/ws/client.go`**
```go
for {
    var raw json.RawMessage
    if err := wsjson.Read(ctx, c.conn, &raw); err != nil {
        return  // connexion fermée ou erreur → on sort de la boucle proprement
    }

    frame, err := parseFrame(raw)
    if err != nil {
        c.sendError("", "", "PARSE_ERROR", "invalid frame")
        continue  // on continue à lire les prochaines frames malgré l'erreur
    }

    switch frame.Type {
    case "message.send":
        var p MessageSendPayload
        json.Unmarshal(frame.Payload, &p)
        view, err := c.messages.SendMessage(ctx, dto.SendMessageCommand{
            RoomID:       model.RoomID(frame.RoomID),
            AuthorID:     userClaims.UserID,
            AuthorName:   userClaims.Username,
            Content:      p.Content,
            ClientTempID: p.ClientTempID,
        })
        // ...
    }
}
```

**④ Le service métier — `Back/application/service/message_service.go`**
```go
func (s *MessageService) SendMessage(ctx context.Context, cmd dto.SendMessageCommand) (dto.MessageView, error) {
    // Validation métier dans le domaine
    msg, err := model.NewMessage(
        model.MessageID(uuid.NewString()),  // UUID généré côté serveur
        cmd.RoomID,
        cmd.AuthorID,
        cmd.Content,      // validé : non vide, ≤ 4000 chars
        cmd.ClientTempID,
    )
    if err != nil {
        return dto.MessageView{}, err  // contenu invalide → remonte l'erreur
    }

    // Stockage
    if err := s.messages.Save(ctx, msg); err != nil {
        return dto.MessageView{}, err
    }

    // Diffusion aux autres membres de la room
    s.broadcastNewMessage(view)

    return view, nil
}
```

---

## 6. Comment les messages sont stockés

### Architecture en deux couches

Les messages ont une double vie : une côté serveur (persistance) et une côté client (affichage).

### Côté serveur — le repository

Dans notre implémentation actuelle, le stockage est **en mémoire** :

```go
// Back/adapter/outbound/memory/message_repo.go

type MessageRepo struct {
    mu       sync.RWMutex    // verrou pour accès concurrent safe (plusieurs goroutines)
    messages []model.Message  // simple slice
}

func (r *MessageRepo) Save(_ context.Context, msg model.Message) error {
    r.mu.Lock()             // on bloque les autres goroutines le temps d'écrire
    defer r.mu.Unlock()
    r.messages = append(r.messages, msg)
    return nil
}

func (r *MessageRepo) FindByRoomID(_ context.Context, roomID model.RoomID, limit int) ([]model.Message, error) {
    r.mu.RLock()            // lecture partagée (plusieurs lecteurs simultanés ok)
    defer r.mu.RUnlock()

    var result []model.Message
    for _, m := range r.messages {
        if m.RoomID == roomID {
            result = append(result, m)
        }
    }

    // Tri chronologique
    sort.Slice(result, func(i, j int) bool {
        return result[i].CreatedAt.Before(result[j].CreatedAt)
    })

    // Limiter aux N derniers messages
    if limit > 0 && len(result) > limit {
        result = result[len(result)-limit:]
    }
    return result, nil
}
```

**Pourquoi l'in-memory ici ?** Pour démarrer vite et tester le protocole WS sans avoir besoin d'une base de données. Les données sont perdues au redémarrage du serveur.

**Migration vers PostgreSQL** : grâce à l'architecture hexagonale, il suffit de créer un `postgres/message_repo.go` qui implémente la même interface `MessageRepository`. Aucune autre ligne ne change dans l'application :

```go
// port/outbound/message_repository.go (l'INTERFACE — ne change JAMAIS)
type MessageRepository interface {
    Save(ctx context.Context, msg model.Message) error
    FindByRoomID(ctx context.Context, roomID model.RoomID, limit int) ([]model.Message, error)
}

// Implémentation mémoire  → utilisée maintenant
// Implémentation postgres → on crée ce fichier plus tard, on swapppe dans main.go
```

### Côté client — le store Zustand

```typescript
// client/src/store/chat.store.ts

interface ChatState {
    messagesByRoom: Record<string, Message[]>  // clé = roomId, valeur = liste de messages
    typingByRoom:   Record<string, TypingUser[]>
}

// Structure en mémoire navigateur :
// {
//   "room-abc-123": [
//     { id: "uuid-1", content: "Salut Bob !", authorName: "alice", pending: false },
//     { id: "uuid-2", content: "Reçu 5/5 !", authorName: "bob",   pending: false },
//   ],
//   "room-def-456": [ ... ]
// }
```

Ce store est **éphémère** : il disparaît au rechargement de page. C'est pour ça qu'au moment de rejoindre une room, le serveur envoie un historique :

```go
// Back/adapter/inbound/ws/client.go — case "room.join"

history, _ := c.messages.GetRoomHistory(ctx, model.RoomID(frame.RoomID), 50)
c.sendJSON(map[string]any{
    "type":    "room.joined",
    "roomId":  frame.RoomID,
    "payload": map[string]any{
        "room":    roomView,
        "history": history,  // ← les 50 derniers messages
    },
})
```

```typescript
// client/src/services/websocket.service.ts — handleStoreUpdates()

case 'room.joined':
    // Peuple le store avec l'historique reçu du serveur
    chat.setHistory(event.roomId!, event.payload.history);
    rooms.updateRoom(event.payload.room);
    break;
```

### Pourquoi Zustand et pas Context API ?

Zustand utilise des **sélecteurs** : un composant ne se re-rend que si la partie du store qu'il observe change.

```typescript
// MessageList ne re-rend QUE quand les messages de CETTE room changent
const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? EMPTY_MESSAGES);
//                                    ^^^^^^^^^^^^^^^^^^^^^^^^^^^
//                                    sélecteur précis = re-render ciblé
```

Avec Context API, n'importe quel changement dans le contexte (même dans une autre room) ferait re-rendre TOUS les composants consommateurs. Avec 100 messages qui arrivent par minute, l'UI serait laggy.

**Piège important** : le sélecteur doit retourner une référence **stable** quand les données ne changent pas. `?? []` est un piège classique :

```typescript
// ❌ Mauvais — crée un nouveau tableau à chaque appel même si rien n'a changé
const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? []);
//                                                              ^^^
//                           [] crée un NOUVEL objet à chaque fois → re-render infini

// ✅ Correct — retourne toujours la même référence quand la room n'a pas de messages
const EMPTY_MESSAGES: Message[] = [];  // constante définie UNE FOIS en dehors du composant
const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? EMPTY_MESSAGES);
```

---

## 7. Concevoir un protocole au-dessus de WebSocket

WebSocket transporte des octets — il ne définit aucune structure pour ton application. Il faut concevoir un protocole applicatif par-dessus.

### Pourquoi une enveloppe standard ?

Sans enveloppe, tu enverrais des messages différents selon le contexte :

```json
// Pour envoyer un message ?
{ "content": "Salut" }

// Pour l'indicateur de frappe ?
{ "username": "alice", "typing": true }

// Pour rejoindre une room ?
{ "action": "join", "room": "general" }
```

C'est impossible à dispatcher proprement. On ne sait pas quoi faire d'un message sans savoir à quel "type" il appartient.

### L'enveloppe de Chuchote

Toutes les frames utilisent une structure uniforme :

```json
{
    "type":      "message.send",
    "requestId": "a1b2c3d4-...",
    "roomId":    "room-xyz",
    "payload":   { ... données spécifiques au type ... }
}
```

- `type` : le discriminant — détermine comment traiter le payload
- `requestId` : UUID généré par le client. Permet au serveur de répondre en référençant la même requête (pour les `ack`, les erreurs).
- `roomId` : scope de l'événement (null pour les événements globaux comme `auth.connect`)
- `payload` : données variables selon le type

### Catalogue complet des événements

**Client → Serveur :**

| Type | Payload | Description |
|---|---|---|
| `auth.connect` | `{ token }` | Authentification. Premier message obligatoire. |
| `room.join` | `{}` | Rejoindre une room (roomId dans l'enveloppe) |
| `room.leave` | `{}` | Quitter une room |
| `message.send` | `{ content, clientTempId }` | Envoyer un message |
| `typing.start` | `{}` | L'utilisateur a commencé à taper |
| `typing.stop` | `{}` | L'utilisateur a arrêté de taper (debounce 2s) |

**Serveur → Client :**

| Type | Payload | Description |
|---|---|---|
| `auth.connected` | `{ userId, username }` | Auth acceptée |
| `auth.error` | `{ code, message }` | Auth refusée, connexion fermée |
| `room.joined` | `{ room, history[] }` | Join accepté + 50 derniers messages |
| `message.new` | `{ id, content, authorName, createdAt }` | Nouveau message dans la room |
| `message.ack` | `{ messageId, clientTempId, createdAt }` | Confirmation de persistance (expéditeur seulement) |
| `typing.indicator` | `{ userId, username, isTyping }` | Broadcast de frappe (sauf à l'émetteur) |
| `user.presence` | `{ userId, status }` | Connexion / déconnexion d'un utilisateur |
| `error` | `{ code, message }` | Erreur générique |

### Le dispatcher côté Go

```go
// Back/adapter/inbound/ws/client.go

switch frame.Type {
case "auth.connect":   // ...
case "room.join":      // ...
case "room.leave":     // ...
case "message.send":   // ...
case "typing.start", "typing.stop": // ...
default:
    slog.Warn("unknown ws event type", "type", frame.Type)
}
```

### Le dispatcher côté TypeScript

```typescript
// client/src/types/ws-events.ts — union discriminée

type InboundWSEvent =
    | AuthConnectedEvent    // type: "auth.connected"
    | AuthErrorEvent        // type: "auth.error"
    | RoomJoinedEvent       // type: "room.joined"
    | MessageNewEvent       // type: "message.new"
    | MessageAckEvent       // type: "message.ack"
    | TypingIndicatorEvent  // type: "typing.indicator"
    | ErrorEvent            // type: "error"

// Le switch est type-safe grâce à l'union discriminée :
// TypeScript sait exactement quels champs sont dans payload selon le type.
switch (event.type) {
    case 'message.new':
        // ici TypeScript SAIT que event.payload est { id, content, authorName, createdAt }
        chat.addMessage(event.roomId!, event.payload);
        break;
}
```

---

## 8. Le pattern UI optimiste

C'est l'une des techniques UX les plus importantes pour les apps temps réel.

### Le problème sans optimisme

```
[Alice frappe + Entrée]
         │
         │ (50 à 200ms de latence réseau)
         │
         ▼
[message affiché]
```

Sur mobile avec une mauvaise connexion, le délai peut dépasser 500ms. L'utilisateur a l'impression que son message n'a pas été envoyé. Il re-clique, envoie en double.

### La solution : afficher d'abord, confirmer ensuite

```
[Alice frappe + Entrée]
         │
         ├──► Affichage IMMÉDIAT (pending: true, bulle grisée)
         │
         │──► Envoi WebSocket en arrière-plan
         │
         │    [50-200ms plus tard]
         │
         ├──◄ message.ack reçu → confirm → bulle devient normale
         │
         └──◄ message.new envoyé à Bob → Bob voit le message
```

### Le clientTempId — la clé du mécanisme

```typescript
// useMessages.ts

const clientTempId = crypto.randomUUID();  // ex: "a1b2c3d4-..."

// Message local immédiat
const optimistic = {
    id: clientTempId,   // id TEMPORAIRE côté client
    content: "Salut Bob !",
    pending: true,      // indicateur visuel
};
addMessage(roomId, optimistic);  // affiché de suite

// Envoyé au serveur avec le même id temporaire
wsService.send({
    type: 'message.send',
    payload: { content: "Salut Bob !", clientTempId },
});
```

```go
// message_service.go — le serveur conserve le clientTempId

msg := model.Message{
    ID:           model.MessageID(uuid.NewString()),  // UUID serveur définitif
    ClientTempID: cmd.ClientTempID,                   // id temporaire du client, conservé
    Content:      cmd.Content,
}
```

```go
// client.go — l'ack est envoyé uniquement à l'expéditeur

c.sendJSON(map[string]any{
    "type":    "message.ack",
    "payload": map[string]any{
        "messageId":    view.ID,           // UUID serveur définitif
        "clientTempId": view.ClientTempID, // id temporaire pour retrouver l'optimiste
        "createdAt":    view.CreatedAt,    // timestamp serveur (référence commune)
    },
})
```

```typescript
// websocket.service.ts — réconciliation

case 'message.ack':
    const existing = chat.messagesByRoom[roomId]?.find(
        (m) => m.clientTempId === clientTempId  // retrouve l'optimiste
    );
    if (existing) {
        chat.confirmMessage(roomId, clientTempId, {
            ...existing,        // garde content, authorName, etc.
            id: messageId,      // remplace l'id temporaire par l'uuid serveur
            createdAt,          // remplace le timestamp local par le timestamp serveur
            pending: false,     // retire l'indicateur visuel
        });
    }
```

### Pourquoi le serveur n'envoie PAS message.new à l'expéditeur

L'expéditeur a déjà l'optimiste dans son UI. Si le serveur lui envoyait aussi un `message.new`, le message apparaîtrait en double. C'est pourquoi on utilise `BroadcastToRoomExcept` :

```go
// message_service.go

func (s *MessageService) broadcastNewMessage(view dto.MessageView) {
    data, _ := json.Marshal(...)

    // Envoie message.new à TOUS les membres SAUF l'expéditeur
    s.hub.BroadcastToRoomExcept(view.RoomID, view.AuthorID, data)
    //                                        ^^^^^^^^^^^^
    //                                        l'expéditeur est exclu
}
```

---

## 9. Gestion des connexions multiples — le Hub

Dans une vraie application, plusieurs utilisateurs sont connectés simultanément. Il faut un mécanisme pour savoir qui est connecté et envoyer des messages aux bonnes personnes.

### Le Hub — architecture

```
                    ┌─────────────────────────────────┐
                    │              Hub                  │
                    │                                   │
  Alice conn-1 ────►│ clients: {                        │
  Alice conn-2 ────►│   conn-1: { userID: alice, send } │
  Bob   conn-3 ────►│   conn-2: { userID: alice, send } │
  Carol conn-4 ────►│   conn-3: { userID: bob,   send } │
                    │   conn-4: { userID: carol, send } │
                    │ }                                 │
                    │                                   │
                    │ roomMembers: {                    │
                    │   "room-general": {conn-1, conn-3}│
                    │   "room-tech":    {conn-2, conn-4}│
                    │ }                                 │
                    │                                   │
                    │ userConns: {                      │
                    │   alice: {conn-1, conn-2}         │ ← Alice sur 2 onglets
                    │   bob:   {conn-3}                 │
                    │   carol: {conn-4}                 │
                    │ }                                 │
                    └─────────────────────────────────┘
```

Alice peut être connectée depuis **plusieurs onglets** simultanément. `userConns` gère ça.

### Thread safety avec sync.RWMutex

Le Hub est accédé par des **dizaines de goroutines en parallèle** (une par connexion WebSocket). Sans protection, des accès concurrents sur les maps Go provoquent une `panic: concurrent map read and map write`.

```go
// Back/adapter/outbound/hub/hub.go

type Hub struct {
    mu          sync.RWMutex  // verrou lecture/écriture
    clients     map[ConnID]client
    roomMembers map[RoomID]map[ConnID]struct{}
    userConns   map[UserID]map[ConnID]struct{}
}

func (h *Hub) BroadcastToRoom(roomID RoomID, payload []byte) {
    h.mu.RLock()   // lecture partagée — plusieurs goroutines peuvent lire en même temps
    defer h.mu.RUnlock()

    for connID := range h.roomMembers[roomID] {
        if c, ok := h.clients[connID]; ok {
            select {
            case c.send <- payload:  // envoie dans le canal sans bloquer
            default:                 // si le canal est plein, on ignore (client trop lent)
            }
        }
    }
}

func (h *Hub) Register(conn Connection, send chan<- []byte) {
    h.mu.Lock()   // écriture exclusive — personne d'autre ne peut lire ni écrire
    defer h.mu.Unlock()
    // ...
}
```

### Le canal `send` par connexion

Chaque client a un canal Go (`chan []byte`). La **goroutine d'écriture** (writePump) lit ce canal et envoie les données sur la WebSocket.

```go
// Back/adapter/inbound/ws/client.go

func (c *Client) run(ctx context.Context) {
    // Lance la goroutine d'écriture en parallèle
    go c.writePump(ctx)

    // La goroutine courante fait la LECTURE
    for {
        // ...lit les messages entrants...
    }
}

func (c *Client) writePump(ctx context.Context) {
    for {
        select {
        case msg, ok := <-c.send:   // attend qu'un message arrive dans le canal
            if !ok {
                return  // canal fermé → connexion terminée
            }
            c.conn.Write(ctx, websocket.MessageText, msg)  // envoie sur la WebSocket
        case <-ctx.Done():
            return  // contexte annulé (connexion fermée)
        }
    }
}
```

Ce pattern sépare proprement **lecture** et **écriture** en deux goroutines indépendantes, évitant tout conflit d'accès concurrent sur la WebSocket.

---

## 10. Cycle de vie complet d'une connexion

```
1. CONNEXION
   Client: new WebSocket("ws://localhost:8080/ws")
   → handshake HTTP/101
   Serveur: websocket.Accept() → crée *websocket.Conn
   → lance newClient().run()
   → lance writePump() en goroutine parallèle

2. AUTHENTIFICATION
   Client: { type: "auth.connect", payload: { token: "JWT..." } }
   Serveur: valide le JWT → extrait userId/username
   → hub.Register(conn, send)  ← connexion enregistrée dans le Hub
   Serveur: { type: "auth.connected", payload: { userId, username } }

3. UTILISATION NORMALE
   Client: { type: "room.join", roomId: "abc" }
   Serveur: → hub.SubscribeToRoom()
   Serveur: { type: "room.joined", payload: { room, history } }

   Client: { type: "message.send", payload: { content } }
   Serveur: → Save → BroadcastToRoomExcept
   Serveur → Alice: { type: "message.ack", ... }
   Serveur → Bob:   { type: "message.new", ... }

4. DÉCONNEXION (fermeture onglet, perte réseau...)
   wsjson.Read() retourne une erreur
   → defer s'exécute: hub.Unregister(conn)
   → supprime la connexion de clients, roomMembers, userConns
   → conn.Close(websocket.StatusNormalClosure, "")

5. RECONNEXION AUTOMATIQUE (côté client)
   ws.onclose = () => {
       setTimeout(() => this.connect(this.currentUrl), 2000)
   }
   → 2 secondes après la déconnexion, retour à l'étape 1
```

---

## 11. WebSocket vs les alternatives

| Technologie | Bidirectionnel | Persistant | Overhead | Complexité | Cas d'usage |
|---|---|---|---|---|---|
| **HTTP Polling** | Non | Non | Très élevé | Faible | À éviter pour du temps réel |
| **Long Polling** | Non | Semi | Élevé | Moyen | Notifications rares, legacy |
| **Server-Sent Events (SSE)** | Non (serveur → client seulement) | Oui | Faible | Faible | Notifications, flux de données en lecture seule |
| **WebSocket** | Oui | Oui | Très faible | Moyen | Chat, jeux, collaboration, trading |
| **WebRTC** | Oui | Oui | Minimal | Élevé | Vidéo/audio P2P, données P2P |

**SSE vs WebSocket** : si tu n'as besoin que d'envoyer des données DU serveur VERS le client (ex : notifications, flux d'actualités), SSE est plus simple à implémenter. WebSocket est nécessaire quand le client doit aussi envoyer des messages (chat, jeux, formulaires collaboratifs).

---

## 12. Ce qu'il faut retenir pour l'oral

### Les 5 points fondamentaux

**1. WebSocket commence par HTTP**
C'est une requête GET ordinaire avec des headers `Upgrade: websocket` et `Connection: Upgrade`. Le serveur répond `101 Switching Protocols`. Après, HTTP est abandonné.

**2. La connexion TCP reste ouverte**
Contrairement à HTTP qui ferme la connexion après chaque échange, WebSocket maintient la connexion TCP active indéfiniment. C'est ce qui permet au serveur de pousser des données à tout moment.

**3. Les frames sont le format de transport**
Les données circulent dans des frames avec un header binaire compact (2 à 10 octets). Les messages client→serveur sont obligatoirement masqués.

**4. WebSocket ne définit pas de protocole applicatif**
Il transporte des octets. C'est à toi de définir ce que tu envoies. En pratique : JSON avec un champ `type` pour distinguer les événements.

**5. Le pattern optimiste est indispensable en prod**
Afficher le message immédiatement côté client, envoyer en arrière-plan, confirmer via un `ack`. Un `clientTempId` permet de faire le lien entre l'optimiste et la confirmation serveur.

### Vocabulaire à maîtriser

- **Handshake** : la négociation HTTP initiale pour établir la connexion WS
- **Frame** : unité de données dans le protocole WebSocket
- **Masquage** : XOR obligatoire sur les messages client→serveur
- **Full-duplex** : les deux parties peuvent envoyer simultanément
- **Hub** : composant serveur qui gère l'ensemble des connexions actives
- **Opcode** : type de frame (texte, binaire, close, ping, pong)
- **UI Optimiste** : afficher avant de confirmer, réconcilier ensuite
- **clientTempId** : identifiant temporaire côté client pour réconcilier l'optimiste avec l'ack

### Schéma de synthèse

```
AVANT websocket (polling)          APRÈS websocket
                                                    
Client ──GET──► Serveur            Client ══════════ Serveur
Client ◄──[]─── Serveur              ↕ permanent ↕  
Client ──GET──► Serveur            push à tout moment
Client ◄──[]─── Serveur              dans les 2 sens
Client ──GET──► Serveur            
Client ◄──data─ Serveur            
                                   
100 req/s pour 100 users           1 connexion par user
500 octets de headers chacune      2-10 octets de header par frame
Latence = intervalle de polling    Latence = RTT réseau seulement
```

---

*Document généré à partir du code source du projet Chuchote — `Back/` (Go) et `client/src/` (React TypeScript)*
