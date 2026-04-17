package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	parkrun "github.com/flopp/parkrun-milestones/internal/parkrun"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Events         []string `yaml:"events"`
	TelegramToken  string   `yaml:"telegram_bot_token"`
	TelegramChatID string   `yaml:"telegram_chat_id"`
	MinActiveRatio float64  `yaml:"min_active_ratio"`
	Runs           uint64   `yaml:"runs"`
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "parkrun-milestones", "config.yaml")
}

func loadConfig(path string) (Config, error) {
	cfg := Config{
		MinActiveRatio: 0.3,
		Runs:           10,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("config-Datei nicht gefunden (%s): %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("ungültige config: %w", err)
	}
	if len(cfg.Events) == 0 {
		return cfg, fmt.Errorf("keine Events in der config-Datei angegeben")
	}
	if cfg.TelegramToken == "" || cfg.TelegramChatID == "" {
		return cfg, fmt.Errorf("telegram_bot_token und telegram_chat_id müssen in der config-Datei gesetzt sein")
	}
	return cfg, nil
}

func sendTelegram(token, chatID, message string) error {
	type payload struct {
		ChatID    string `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}
	body, err := json.Marshal(payload{ChatID: chatID, Text: message, ParseMode: "HTML"})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API Fehler: %s – %s", resp.Status, string(respBody))
	}
	return nil
}

func milestoneLabel(current int64) string {
	return fmt.Sprintf("%d → <b>%d</b> ⭐", current, current+1)
}

func buildMessage(event *parkrun.Event, parkrunners []*parkrun.Parkrunner, examinedRuns uint64) string {
	junior := event.IsJuniorParkrun()
	nextRun := len(event.Runs) + 1

	var milestones []string
	for _, p := range parkrunners {
		var parts []string
		if junior {
			if parkrun.Milestone(p.JuniorRuns + 1) {
				parts = append(parts, "Läufe: "+milestoneLabel(p.JuniorRuns))
			}
		} else {
			if parkrun.Milestone(p.Runs + 1) {
				parts = append(parts, "Läufe: "+milestoneLabel(p.Runs))
			}
		}
		if parkrun.Milestone(p.Vols + 1) {
			parts = append(parts, "Ehrenamt: "+milestoneLabel(p.Vols))
		}
		if len(parts) > 0 {
			milestones = append(milestones, fmt.Sprintf("🎯 %s – %s", html.EscapeString(p.Name), strings.Join(parts, ", ")))
		}
	}

	header := fmt.Sprintf("🏃 <b>Milestone-Vorschau: %s – Lauf #%d</b>", html.EscapeString(event.Name), nextRun)
	if len(milestones) == 0 {
		return header + "\n\nKeine Meilensteine diese Woche."
	}
	return header + "\n\n" + strings.Join(milestones, "\n") +
		fmt.Sprintf("\n\n<i>Aktivität: letzte %d Läufe berücksichtigt</i>", examinedRuns)
}

func main() {
	configPath := flag.String("config", defaultConfigPath(), "Pfad zur config.yaml")
	forceReload := flag.Bool("force", false, "Alle gecachten Daten neu laden")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fehler:", err)
		os.Exit(1)
	}

	if *forceReload {
		parkrun.MaxFileAge = 0
	}

	for _, eventId := range cfg.Events {
		event, err := parkrun.LookupEvent(eventId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Event %q: %v\n", eventId, err)
			continue
		}

		fmt.Printf("Lade Daten für %s...\n", event.Name)
		parkrunners, examinedRuns, err := event.GetActiveParkrunners(cfg.MinActiveRatio, cfg.Runs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Event %q: %v\n", eventId, err)
			continue
		}

		msg := buildMessage(event, parkrunners, examinedRuns)
		fmt.Println(msg)

		if err := sendTelegram(cfg.TelegramToken, cfg.TelegramChatID, msg); err != nil {
			fmt.Fprintf(os.Stderr, "Telegram-Fehler: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Telegram-Nachricht gesendet.")
	}
}
