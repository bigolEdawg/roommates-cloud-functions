# 🏠 Roommate App — Groups & Invites (GCP MVP)

This is a concrete starter plan to implement **Groups & Invites** on Google Cloud (GCP) using **Firebase Auth + Firestore + Cloud Functions (Gen 2)**. It includes architecture, data model, API contracts, security rules, and Go function stubs.

---

## 0) Tech Stack (MVP)
- **Auth**: Firebase Auth (email/password, Apple/Google later)
- **DB**: Firestore (Native mode)
- **Backend**: Cloud Functions (Gen 2, **Go**) — package name `function`
- **Tokens**: Short-lived invite tokens stored in Firestore
- **Notifications (optional)**: Firebase Cloud Messaging
- **Linking**: Firebase Dynamic Links for deep links to accept invites in-app

---

## 1) Firestore Data Model

### Collections & Docs
```
/groups/{groupId}
  name: string
  createdAt: timestamp
  createdBy: uid
  settings: {
    houseRules: string,           // freeform markdown for now
    quietHours: { start: string, end: string },
  }
  stats: {
    memberCount: number,
    inviteCount: number
  }

/groups/{groupId}/members/{uid}
  role: "owner" | "member"
  joinedAt: timestamp

/invites/{inviteId}
  groupId: string
  emailOrPhone: string            // email for MVP
  token: string                   // random URL-safe ID
  status: "pending" | "accepted" | "expired" | "revoked"
  createdAt: timestamp
  createdBy: uid
  expiresAt: timestamp            // e.g., now + 7 days

/users/{uid}
  displayName: string
  email: string
  groups: string[]                // list of groupIds for quick lookups

/users/{uid}/pendingInvites/{inviteId}
  groupId: string
  status: string
  createdAt: timestamp
```

**Indexes** (Composite)
- `/invites` on `(token ASC)` **or** single-field token index (exact lookups)
- `/invites` on `(emailOrPhone ASC, status ASC)` to list pending invites by email (optional)
- `/groups/{groupId}/members` on `(role ASC)` if you filter by role (optional)

---

## 2) Security Rules (Draft)
> Lock down by default; owners manage invites; members can read group + members; only the invite target can accept.

```rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    // Users can read their own user doc & pending invites
    match /users/{uid} {
      allow read, write: if request.auth != null && request.auth.uid == uid;

      match /pendingInvites/{inviteId} {
        allow read, write: if request.auth != null && request.auth.uid == uid;
      }
    }

    // Groups
    match /groups/{groupId} {
      allow read: if isMember(groupId);
      allow create: if request.auth != null; // anyone authenticated can create
      allow update, delete: if isOwner(groupId);

      match /members/{memberUid} {
        allow read: if isMember(groupId);
        allow create, update, delete: if isOwner(groupId);
      }
    }

    // Invites — read by owners; accept via Cloud Function (server-side)
    match /invites/{inviteId} {
      allow read: if isOwner(resource.data.groupId);
      allow create: if isOwner(request.resource.data.groupId);
      allow update, delete: if isOwner(resource.data.groupId);
    }

    function isMember(groupId) {
      return request.auth != null &&
             exists(/databases/$(database)/documents/groups/$(groupId)/members/$(request.auth.uid));
    }

    function isOwner(groupId) {
      return request.auth != null &&
        get(/databases/$(database)/documents/groups/$(groupId)/members/$(request.auth.uid)).data.role == 'owner';
    }
  }
}
```

> Note: Accepting an invite should be done **server-side** (via Cloud Function) because it writes to `/groups/{groupId}/members` and updates `/invites/{inviteId}`; do not expose this directly to a client to avoid privilege escalation.

---

## 3) API Contracts (HTTP Cloud Functions)

Base path example (you’ll deploy each as its own function):
- `POST /groups.create` — Create a group
- `POST /invites.create` — Create an invite (owner only)
- `POST /invites.accept` — Accept an invite by token (authenticated)
- `POST /invites.revoke` — Revoke a pending invite (owner only)
- `GET  /groups.listMine` — List groups for current user

### 3.1 `POST /groups.create`
**Auth:** required
**Body:** `{ name: string, settings?: { houseRules?: string, quietHours?: {start, end} } }`
**Returns:** `{ groupId, name }`
**Side effects:**
- Creates `/groups/{groupId}`
- Creates `/groups/{groupId}/members/{uid}` with role `owner`
- Adds groupId to `/users/{uid}.groups`

### 3.2 `POST /invites.create`
**Auth:** required (owner)
**Body:** `{ groupId: string, email: string, expiresInDays?: number }`
**Returns:** `{ inviteId, token, deepLink }`
**Notes:**
- Generates a random `token`
- Writes `/invites/{inviteId}` with `expiresAt`
- (Optional) mirror to `/users/{uid}/pendingInvites/{inviteId}` if target already has an account
- Build a Firebase Dynamic Link `myapp://invite?token=...` or web fallback

### 3.3 `POST /invites.accept`
**Auth:** required
**Body:** `{ token: string }`
**Returns:** `{ groupId }`
**Server logic:**
- Lookup invite by `token`
- Validate `status == 'pending'` and `now < expiresAt`
- Add `request.auth.uid` to `/groups/{groupId}/members/{uid}` as `member`
- Update `/invites/{inviteId}.status = 'accepted'`
- Add groupId to `/users/{uid}.groups`

### 3.4 `POST /invites.revoke`
**Auth:** required (owner)
**Body:** `{ inviteId: string }`
**Returns:** `{ ok: true }`

### 3.5 `GET /groups.listMine`
**Auth:** required
**Returns:** `[{ groupId, name, role }]`

---

## 4) Go — Cloud Functions (Gen 2) Stubs
> Package name `function`; minimal error handling for clarity. Use Firestore Admin SDK.

```go
// go.mod (excerpt)
// module yourdomain/roommate
// go 1.22
// require (
//   cloud.google.com/go/firestore v1.14.0
//   firebase.google.com/go/v4 v4.13.0
//   github.com/google/uuid v1.5.0
// )
```

```go
// file: groups.go
package function

import (
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "os"
  "time"

  fb "firebase.google.com/go/v4"
  "cloud.google.com/go/firestore"
  "google.golang.org/api/option"
)

type GroupCreateReq struct {
  Name     string `json:"name"`
  Settings struct {
    HouseRules string `json:"houseRules"`
    QuietHours struct { Start string `json:"start"`; End string `json:"end"` } `json:"quietHours"`
  } `json:"settings"`
}

type authContext struct{ UID string }

func getAuth(r *http.Request) (authContext, error) {
  // Expect Firebase Auth ID token via Authorization: Bearer <token>
  // In Gen 2, prefer Identity Platform / Firebase Admin token verification.
  // For brevity, assume reverse proxy injects X-UID (dev only!). Replace with real verification.
  uid := r.Header.Get("X-UID")
  if uid == "" { return authContext{}, fmt.Errorf("unauthorized") }
  return authContext{UID: uid}, nil
}

func firestoreClient(ctx context.Context) (*firestore.Client, error) {
  app, err := fb.NewApp(ctx, nil, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
  if err != nil { return nil, err }
  return app.Firestore(ctx)
}

func GroupsCreate(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  auth, err := getAuth(r)
  if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

  var req GroupCreateReq
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
  if req.Name == "" { http.Error(w, "name required", 400); return }

  fs, err := firestoreClient(ctx); if err != nil { http.Error(w, err.Error(), 500); return }
  defer fs.Close()

  grp := map[string]interface{}{
    "name": req.Name,
    "createdAt": firestore.ServerTimestamp,
    "createdBy": auth.UID,
    "settings": map[string]interface{}{
      "houseRules": req.Settings.HouseRules,
      "quietHours": map[string]interface{}{"start": req.Settings.QuietHours.Start, "end": req.Settings.QuietHours.End},
    },
    "stats": map[string]interface{}{"memberCount": 1, "inviteCount": 0},
  }

  ref, _, err := fs.Collection("groups").Add(ctx, grp)
  if err != nil { http.Error(w, err.Error(), 500); return }

  // add owner membership
  _, err = fs.Collection("groups").Doc(ref.ID).Collection("members").Doc(auth.UID).Set(ctx, map[string]interface{}{
    "role": "owner",
    "joinedAt": firestore.ServerTimestamp,
  })
  if err != nil { http.Error(w, err.Error(), 500); return }

  // append to user.groups
  userRef := fs.Collection("users").Doc(auth.UID)
  _, _ = userRef.Set(ctx, map[string]interface{}{"groups": firestore.ArrayUnion(ref.ID)}, firestore.MergeAll)

  json.NewEncoder(w).Encode(map[string]string{"groupId": ref.ID, "name": req.Name})
}
```

```go
// file: invites.go
package function

import (
  "context"
  "encoding/json"
  "net/http"
  "time"

  "github.com/google/uuid"
  "cloud.google.com/go/firestore"
)

type InviteCreateReq struct {
  GroupID       string `json:"groupId"`
  Email         string `json:"email"`
  ExpiresInDays int    `json:"expiresInDays"`
}

type InviteAcceptReq struct {
  Token string `json:"token"`
}

func InvitesCreate(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  auth, err := getAuth(r)
  if err != nil { http.Error(w, "unauthorized", 401); return }

  var req InviteCreateReq
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
  if req.GroupID == "" || req.Email == "" { http.Error(w, "groupId and email required", 400); return }
  if req.ExpiresInDays <= 0 { req.ExpiresInDays = 7 }

  fs, err := firestoreClient(ctx); if err != nil { http.Error(w, err.Error(), 500); return }
  defer fs.Close()

  // Verify owner
  mem, err := fs.Collection("groups").Doc(req.GroupID).Collection("members").Doc(auth.UID).Get(ctx)
  if err != nil || mem.Data()["role"] != "owner" { http.Error(w, "forbidden", 403); return }

  token := uuid.NewString()
  expires := time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour)

  inv := map[string]interface{}{
    "groupId": req.GroupID,
    "emailOrPhone": req.Email,
    "token": token,
    "status": "pending",
    "createdAt": firestore.ServerTimestamp,
    "createdBy": auth.UID,
    "expiresAt": expires,
  }

  ref, _, err := fs.Collection("invites").Add(ctx, inv)
  if err != nil { http.Error(w, err.Error(), 500); return }

  // (Optional) increment stats.inviteCount
  _, _ = fs.Collection("groups").Doc(req.GroupID).Set(ctx, map[string]interface{}{
    "stats": map[string]interface{}{"inviteCount": firestore.Increment(1)},
  }, firestore.MergeAll)

  deepLink := "myapp://invite?token=" + token // replace later with Firebase Dynamic Links
  json.NewEncoder(w).Encode(map[string]string{"inviteId": ref.ID, "token": token, "deepLink": deepLink})
}

func InvitesAccept(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  auth, err := getAuth(r)
  if err != nil { http.Error(w, "unauthorized", 401); return }

  var req InviteAcceptReq
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
  if req.Token == "" { http.Error(w, "token required", 400); return }

  fs, err := firestoreClient(ctx); if err != nil { http.Error(w, err.Error(), 500); return }
  defer fs.Close()

  // Lookup invite by token
  qs, err := fs.Collection("invites").Where("token", "==", req.Token).Limit(1).Documents(ctx).GetAll()
  if err != nil || len(qs) == 0 { http.Error(w, "invalid token", 404); return }

  inv := qs[0]
  data := inv.Data()

  if status, _ := data["status"].(string); status != "pending" { http.Error(w, "invite not pending", 409); return }
  if exp, ok := data["expiresAt"].(time.Time); ok && time.Now().After(exp) { http.Error(w, "invite expired", 410); return }

  groupId, _ := data["groupId"].(string)
  if groupId == "" { http.Error(w, "corrupt invite", 500); return }

  // Add membership as member
  _, err = fs.Collection("groups").Doc(groupId).Collection("members").Doc(auth.UID).Set(ctx, map[string]interface{}{
    "role": "member",
    "joinedAt": firestore.ServerTimestamp,
  }, firestore.MergeAll)
  if err != nil { http.Error(w, err.Error(), 500); return }

  // Update invite
  _, _ = inv.Ref.Set(ctx, map[string]interface{}{"status": "accepted"}, firestore.MergeAll)

  // Update user doc
  _, _ = fs.Collection("users").Doc(auth.UID).Set(ctx, map[string]interface{}{
    "groups": firestore.ArrayUnion(groupId),
  }, firestore.MergeAll)

  json.NewEncoder(w).Encode(map[string]string{"groupId": groupId})
}
```

---

## 5) Deployment (Gen 2 — HTTP)
> Example using `gcloud` (replace region/project/function names). Ensure proper IAM and service account permissions for Firestore.

```bash
# Build
gcloud functions deploy groups-create \
  --gen2 --region=us-west2 --runtime=go122 \
  --entry-point=GroupsCreate --trigger-http --allow-unauthenticated=false

gcloud functions deploy invites-create \
  --gen2 --region=us-west2 --runtime=go122 \
  --entry-point=InvitesCreate --trigger-http --allow-unauthenticated=false

gcloud functions deploy invites-accept \
  --gen2 --region=us-west2 --runtime=go122 \
  --entry-point=InvitesAccept --trigger-http --allow-unauthenticated=false
```

> For local dev you can temporarily pass `X-UID: <test-uid>` header as seen in the stubs, but **replace with real Firebase ID token verification** before launch.

---

## 6) Client Flows (Mobile UI)

### Create Group
1) User taps **Create Group** → enters name → POST `/groups.create`
2) On success, navigate to Group Home (tabs: Groceries, Chores, Bills, Rules)

### Invite Member (Email)
1) Owner opens **Members** → **Invite** → enters email → POST `/invites.create`
2) Show share sheet with `deepLink` (or auto-send email via SendGrid later)
3) Invitee opens the link → app captures `token` → if not logged in, sign in → POST `/invites.accept`

### Accept Invite (Deep Link)
- App registered for `myapp://invite?token=...`
- After login, call `/invites.accept`
- On success, store `groupId` in Redux/Local state and route to Group Home.

---

## 7) Observability & Admin
- **Logs**: Cloud Logging with structured fields (groupId, inviteId, uid)
- **Metrics**: Count of created/accepted/expired invites, members per group
- **Admin**: A simple `/admin/revokeInvite` function (owner only) if needed

---

## 8) Nice-to-Haves (Short Horizon)
- Expire invites hourly via **Cloud Scheduler** (set `status='expired'` when `expiresAt < now`)
- Email sender via **SendGrid** or **Mailgun**
- Phone/SMS invites via **Twilio** (store `emailOrPhone` type)
- Role upgrades/downgrades (admin vs member)

---

## 9) Test Plan (Quick)
- Create group → verify owner membership, stats.memberCount=1
- Create invite (owner) → verify `/invites` doc and deepLink
- Accept invite (member) → verify member doc, invite status=accepted, user.groups updated
- Try accept twice → should 409 (not pending)
- Try accept after expiration → should 410
- Non-owner create invite → 403

---

## 10) Curl Examples (Dev)
```bash
# Create group
curl -X POST $GROUPS_CREATE \
  -H "Content-Type: application/json" -H "X-UID: test-owner-123" \
  -d '{"name":"House on 42nd","settings":{"houseRules":"Be kind","quietHours":{"start":"22:00","end":"07:00"}}}'

# Create invite
curl -X POST $INVITES_CREATE \
  -H "Content-Type: application/json" -H "X-UID: test-owner-123" \
  -d '{"groupId":"<GROUP_ID>","email":"friend@example.com","expiresInDays":7}'

# Accept invite
curl -X POST $INVITES_ACCEPT \
  -H "Content-Type: application/json" -H "X-UID: test-member-456" \
  -d '{"token":"<TOKEN_FROM_CREATE>"}'
```

---

## 11) TODO Checklist (You Can Work Top → Bottom)
- [ ] Initialize Firebase project (Auth, Firestore)
- [ ] Add Firestore rules from this spec
- [ ] Create Go repo with `function` package
- [ ] Implement `GroupsCreate`, `InvitesCreate`, `InvitesAccept`
- [ ] Deploy 3 functions (Gen 2)
- [ ] Register app deep link (e.g., `myapp://`)
- [ ] Wire up mobile: create group → invite → accept
- [ ] Add basic Members screen (list members, roles)
- [ ] Add Cloud Scheduler (expire old invites)
- [ ] Replace `X-UID` with real Firebase token verification

---

**You now have a fully-scoped MVP for Groups & Invites on GCP, with code stubs and deployment commands.** Next step: set up the Firebase project and deploy `groups-create` first.

