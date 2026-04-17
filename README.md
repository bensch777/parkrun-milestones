# parkrun-milestones

Sendet jeden Freitag eine Telegram-Nachricht mit den Milestone-Kandidaten für den nächsten [Krupunder See parkrun](https://www.parkrun.com.de/krupundersee/).

Als Milestone-Kandidat gilt, wer beim nächsten Lauf eine Jubiläumszahl erreicht (25, 50, 100, 150, 200, ...) und in mindestens 30% der letzten 10 Läufe aktiv war.

## Einrichtung

### 1. Binary bauen

```bash
make build
```

### 2. Config anlegen

```bash
mkdir -p ~/.config/parkrun-milestones
cp config.yaml.example ~/.config/parkrun-milestones/config.yaml
```

Dann `~/.config/parkrun-milestones/config.yaml` bearbeiten und Telegram-Zugangsdaten eintragen:

```yaml
events:
  - krupundersee

telegram_bot_token: "123456789:ABC..."
telegram_chat_id: "-987654321"
```

> **Telegram-Bot erstellen:** Bei `@BotFather` mit `/newbot` → Token kopieren.  
> **Chat-ID einer Gruppe:** Bot in die Gruppe einladen, dann Chat-URL prüfen (`#-XXXXXXXXX`).

### 3. Automatisch freitags ausführen (macOS LaunchAgent)

Der LaunchAgent läuft jeden Freitag um 20:00 Uhr und ist unter
`~/Library/LaunchAgents/com.bensch.parkrun-milestones.plist` eingerichtet.

## Manuell ausführen

```bash
./.bin/parkrun-milestones
```

Mit `-force` werden alle gecachten Daten neu geladen:

```bash
./.bin/parkrun-milestones -force
```

## Beispiel-Nachricht

```
🏃 Milestone-Vorschau: Krupunder See parkrun – Lauf #63

🎯 Ben RATHMANN – Läufe: 24 → 25 ⭐
🎯 Benjamin NESKE – Läufe: 24 → 25 ⭐

Aktivität: letzte 10 Läufe berücksichtigt
```
