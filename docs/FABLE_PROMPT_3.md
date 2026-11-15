# Pass 3: replace the login screen

The user rejected the account name + password form. The Go backend no longer
has `LoginWithPassword` or `SubmitGuardCode`; the LoginState phases
`submitting` and `guard` and the fields `guards`, `emailHint`, `codeError`,
`checking` are gone. New: `LocalLogins`, `UseSteamAccount`, `OpenSteamLogin`,
event `mychat:local-logins`, `Snapshot.local`, `Status.source`.

Re-read `docs/UI_CONTRACT.md` (section "Settings window rules → Login screen")
and the regenerated `frontend/bindings/`. Work only in `frontend/`. Do not edit
Go files or `frontend/bindings/`. Keep everything else from pass 2 (overlay,
avatars, animations, other tabs) as it is.

## What to build

1. Remove the password form, the Steam Guard step and all code for them
   (components, store fields, mock logic, trust line about passwords).
2. New ACCOUNT login screen in the same Panorama style, with three ways in
   this order:
   - **Use Steam on this PC** (default, the main element): account cards from
     `local.accounts`, styled like the CS2 spectator player cards (circle
     letter avatar with ring, persona name in condensed caps, account name
     muted, gold highlight on the selected / most recent card). One primary
     button `USE THIS ACCOUNT`. Cards without `hasLogin` are muted with
     `SIGN IN IN STEAM` (→ `OpenSteamLogin()`).
   - **Sign in with the Steam app**: `OPEN STEAM SIGN-IN` → `OpenSteamLogin()`,
     and the waiting state while `local.watching`.
   - **QR code**: collapsed by default. Expanding calls `StartLogin()`,
     collapsing calls `CancelLogin()`. Keep the existing QR rendering.
   - No Steam installed (`steamInstalled: false`) → only the QR, expanded.
   - `login.error` on `idle` → a banner at the top (e.g. expired login).
3. Logged-in ACCOUNT view: show `source` ("Using Steam on this PC" or "QR
   login"), `Log out`, and for `source: "steam"` a `Switch account` button
   that shows the account cards again (picking one calls `UseSteamAccount`).
4. Handle `mychat:local-logins` in the store; take the initial value from
   `Snapshot.local`.

## Mock mode

- `?view=settings&mock=login`: three accounts, two with `hasLogin` (one most
  recent), one without. `USE THIS ACCOUNT` goes online after ~700 ms.
  An account named `expired` rejects with "Steam has no saved login for
  expired. Sign in to Steam with \"Remember me\" on".
- `mock=watching`: the waiting-for-Steam state; after ~3 s a new account
  appears and the mock goes online.
- `mock=nosteam`: Steam not installed → only QR.
- `mock=expired`: login screen with the expired banner.
- Remove `mock=guard`; `mock=error` should now show a failed
  `UseSteamAccount`.

## Finish

`npm run check` and `npm run build` must pass with zero errors and zero
warnings. Reply with a short summary and the list of mock URLs.
