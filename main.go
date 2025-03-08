package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type TimeZoneDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type AvailabilityRequest struct {
	ServiceID     string           `json:"serviceId"`
	StaffIDs      []string         `json:"staffIds"`
	StartDateTime TimeZoneDateTime `json:"startDateTime"`
	EndDateTime   TimeZoneDateTime `json:"endDateTime"`
}

type AvailabilityItem struct {
	Status        string `json:"status"`
	StartDateTime struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"startDateTime"`
	EndDateTime struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"endDateTime"`
	AvailableCount int `json:"availableCount"`
}

type AvailabilityResponse struct {
	StaffAvailabilityResponse []struct {
		StaffId           string             `json:"staffId"`
		AvailabilityItems []AvailabilityItem `json:"availabilityItems"`
	} `json:"staffAvailabilityResponse"`
}

// MessageSender is an interface for sending messages
type MessageSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type AppointmentChecker struct {
	bot               *tgbotapi.BotAPI
	client            *http.Client
	checkInterval     time.Duration
	users             map[int64]bool // All users that ever interacted with bot
	autoCheckUsers    map[int64]bool // Users who want periodic status updates
	usersMutex        sync.RWMutex
	lastAvailableSlot *AvailabilityItem
	wasAvailable      bool // Track previous availability state
	logger            *log.Logger
}

func setupLogger() (*log.Logger, error) {
	// Create logger with timestamp and caller info
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.Printf("Bot logging started")
	return logger, nil
}

func NewAppointmentChecker(bot *tgbotapi.BotAPI) *AppointmentChecker {
	logger, err := setupLogger()
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	ac := &AppointmentChecker{
		bot:            bot,
		client:         &http.Client{Timeout: 30 * time.Second},
		checkInterval:  1 * time.Minute, // Check every minute for changes
		users:          make(map[int64]bool),
		autoCheckUsers: make(map[int64]bool),
		logger:         logger,
	}

	if bot != nil {
		ac.logger.Printf("Bot initialized with username: @%s", bot.Self.UserName)
	}
	return ac
}

// AvailabilityResult represents the result of an availability check
type AvailabilityResult struct {
	Available     bool
	AvailableSlot *AvailabilityItem
}

func (ac *AppointmentChecker) logUserAction(chatID int64, action, details string) {
	ac.logger.Printf("User %d: %s - %s", chatID, action, details)
}

func (ac *AppointmentChecker) checkAvailability() (AvailabilityResult, error) {
	ac.logger.Printf("Starting availability check")
	startTime := time.Now()
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 2, 0)

	// First appointment type
	req1 := AvailabilityRequest{
		ServiceID: "1df7f565-8337-412b-91ec-b8ffd49fe6f2",
		StaffIDs:  []string{"4f3b2516-99cd-4295-9328-afefb3b403e3"},
		StartDateTime: TimeZoneDateTime{
			DateTime: startDate.Format("2006-01-02T15:04:05"),
			TimeZone: "FLE Standard Time",
		},
		EndDateTime: TimeZoneDateTime{
			DateTime: endDate.Format("2006-01-02T15:04:05"),
			TimeZone: "FLE Standard Time",
		},
	}

	// Second appointment type
	req2 := AvailabilityRequest{
		ServiceID: "51b3c1e4-2dc8-46ab-88e3-604cb4164c4c",
		StaffIDs:  []string{"84d3f0dd-33f9-4d2d-a741-98b86e790315"},
		StartDateTime: TimeZoneDateTime{
			DateTime: startDate.Format("2006-01-02T15:04:05"),
			TimeZone: "FLE Standard Time",
		},
		EndDateTime: TimeZoneDateTime{
			DateTime: endDate.Format("2006-01-02T15:04:05"),
			TimeZone: "FLE Standard Time",
		},
	}

	available1, err := ac.checkEndpoint(req1)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("error checking first endpoint: %v", err)
	}

	available2, err := ac.checkEndpoint(req2)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("error checking second endpoint: %v", err)
	}

	result := AvailabilityResult{
		Available:     available1 || available2,
		AvailableSlot: ac.lastAvailableSlot,
	}

	ac.logger.Printf("Availability check completed in %v. Available: %v", time.Since(startTime), result.Available)
	return result, nil
}

func (ac *AppointmentChecker) checkEndpoint(req AvailabilityRequest) (bool, error) {
	ac.logger.Printf("Checking endpoint for service ID: %s", req.ServiceID)
	startTime := time.Now()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	url := "https://outlook.office365.com/BookingsService/api/V1/bookingBusinessesc2/monetrapirkanmaarekrytointipalvelut@monetra.fi/GetStaffAvailability?app=BookingsC1"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := ac.client.Do(request)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var response AvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %v", err)
	}

	// Check each staff member's availability
	for _, staff := range response.StaffAvailabilityResponse {
		for _, item := range staff.AvailabilityItems {
			// Check if the slot is either fully available or has available slots
			if item.Status == "BOOKINGSAVAILABILITYSTATUS_AVAILABLE" ||
				item.Status == "BOOKINGSAVAILABILITYSTATUS_SLOTS_AVAILABLE" {
				if item.AvailableCount > 0 {
					startTime, err := time.Parse(time.RFC3339, item.StartDateTime.DateTime)
					if err != nil {
						continue
					}
					// Only consider future slots
					if startTime.After(time.Now()) {
						ac.lastAvailableSlot = &item
						ac.logger.Printf("Found available slot at %s with %d slots", startTime.Format("2006-01-02 15:04:05"), item.AvailableCount)
						return true, nil
					}
				}
			}
		}
	}

	ac.logger.Printf("Endpoint check completed in %v. No available slots found", time.Since(startTime))
	return false, nil
}

func (ac *AppointmentChecker) getMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Check Availability"),
			tgbotapi.NewKeyboardButton("📊 Status"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⏰ Enable Status Updates"),
			tgbotapi.NewKeyboardButton("⏳ Disable Status Updates"),
		),
	)
}

func (ac *AppointmentChecker) handleCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	username := update.Message.From.UserName
	if username == "" {
		username = fmt.Sprintf("id:%d", update.Message.From.ID)
	}

	// Log incoming message
	ac.logger.Printf("Received message from @%s (ID: %d): %s", username, chatID, update.Message.Text)

	// Add user to the users map
	ac.usersMutex.Lock()
	ac.users[chatID] = true
	ac.usersMutex.Unlock()

	var command string

	// Handle both text messages and commands
	if update.Message.IsCommand() {
		command = update.Message.Command()
	} else {
		// Map button texts to commands
		switch update.Message.Text {
		case "🔍 Check Availability":
			command = "check"
		case "📊 Status":
			command = "status"
		case "⏰ Enable Status Updates":
			command = "autostart"
		case "⏳ Disable Status Updates":
			command = "autostop"
		default:
			command = "help"
		}
	}

	keyboard := ac.getMainKeyboard()

	switch command {
	case "start", "help":
		ac.logUserAction(chatID, "COMMAND", "Started bot/requested help")
		msg := tgbotapi.NewMessage(chatID, "👋 Welcome! I'll notify you when appointment availability changes.\n\n"+
			"Available commands:\n"+
			"🔍 Check Availability - Check current appointment availability\n"+
			"📊 Status - Show your notification settings\n"+
			"⏰ Enable Status Updates - Get status update every 30 minutes\n"+
			"⏳ Disable Status Updates - Only get notifications when availability changes")
		msg.ReplyMarkup = keyboard
		ac.bot.Send(msg)

	case "autostart":
		ac.logUserAction(chatID, "AUTO-CHECK", "Enabled status updates")
		ac.usersMutex.Lock()
		ac.autoCheckUsers[chatID] = true
		ac.usersMutex.Unlock()

		msg := tgbotapi.NewMessage(chatID, "⏰ Status updates enabled! You'll receive availability updates every 30 minutes.")
		msg.ReplyMarkup = keyboard
		ac.bot.Send(msg)

	case "autostop":
		ac.logUserAction(chatID, "AUTO-CHECK", "Disabled status updates")
		ac.usersMutex.Lock()
		delete(ac.autoCheckUsers, chatID)
		ac.usersMutex.Unlock()

		msg := tgbotapi.NewMessage(chatID, "⏳ Status updates disabled. You'll only be notified when appointment availability changes.")
		msg.ReplyMarkup = keyboard
		ac.bot.Send(msg)

	case "check":
		ac.logUserAction(chatID, "CHECK", "Manual availability check")
		msg := tgbotapi.NewMessage(chatID, "🔍 Checking appointment availability...")
		msg.ReplyMarkup = keyboard
		ac.bot.Send(msg)

		result, err := ac.checkAvailability()
		if err != nil {
			errMsg := fmt.Sprintf("❌ Error checking availability: %v", err)
			reply := tgbotapi.NewMessage(chatID, errMsg)
			reply.ReplyMarkup = keyboard
			ac.bot.Send(reply)
			return
		}

		checkMsg := ac.formatAvailabilityMessage(result, true)
		reply := tgbotapi.NewMessage(chatID, checkMsg)
		reply.ReplyMarkup = keyboard
		ac.bot.Send(reply)

	case "status":
		ac.logUserAction(chatID, "STATUS", "Checked status")
		ac.usersMutex.RLock()
		autoCheck := ac.autoCheckUsers[chatID]
		ac.usersMutex.RUnlock()

		var status string
		if autoCheck {
			status = "🟢 You will receive availability updates every 30 minutes."
		} else {
			status = "🔵 You will be notified only when appointment availability changes."
		}

		msg := tgbotapi.NewMessage(chatID, status)
		msg.ReplyMarkup = keyboard
		ac.bot.Send(msg)
	}
}

func (ac *AppointmentChecker) formatAvailabilityMessage(result AvailabilityResult, isManualCheck bool) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	if result.Available {
		if result.AvailableSlot != nil {
			startTime, _ := time.Parse(time.RFC3339, result.AvailableSlot.StartDateTime.DateTime)
			return fmt.Sprintf("🎉 Appointments are available!\n\nNext available appointment:\nDate: %s\nTime: %s\nAvailable slots: %d\n\nBooking website: https://outlook.office365.com/owa/calendar/monetrapirkanmaarekrytointipalvelut@monetra.fi/bookings/",
				startTime.Format("Monday, January 2, 2006"),
				startTime.Format("15:04"),
				result.AvailableSlot.AvailableCount)
		}
		return "🎉 Appointments are available!\n\nBooking website: https://outlook.office365.com/owa/calendar/monetrapirkanmaarekrytointipalvelut@monetra.fi/bookings/"
	}

	if isManualCheck {
		return fmt.Sprintf("❌ No appointments available (checked at %s)", currentTime)
	}
	return fmt.Sprintf("❌ Appointments are no longer available (as of %s)", currentTime)
}

func (ac *AppointmentChecker) notifyUsers(result AvailabilityResult, isAutoCheck bool) {
	ac.usersMutex.RLock()
	defer ac.usersMutex.RUnlock()

	// For availability changes, notify all users
	if result.Available != ac.wasAvailable {
		for chatID := range ac.users {
			msg := ac.formatAvailabilityMessage(result, false)
			reply := tgbotapi.NewMessage(chatID, msg)
			reply.ReplyMarkup = ac.getMainKeyboard()
			ac.bot.Send(reply)
		}
	}

	// For auto-check users, send periodic updates if enabled
	if isAutoCheck {
		for chatID := range ac.autoCheckUsers {
			msg := ac.formatAvailabilityMessage(result, true)
			reply := tgbotapi.NewMessage(chatID, msg)
			reply.ReplyMarkup = ac.getMainKeyboard()
			ac.bot.Send(reply)
		}
	}
}

func (ac *AppointmentChecker) runBot() {
	ac.logger.Printf("Bot started with check interval: %v", ac.checkInterval)

	// Start the availability checker in a separate goroutine
	go func() {
		// Ticker for regular availability checks
		checkTicker := time.NewTicker(ac.checkInterval)
		// Ticker for auto-check updates (30 minutes)
		autoCheckTicker := time.NewTicker(30 * time.Minute)
		defer checkTicker.Stop()
		defer autoCheckTicker.Stop()

		checkCount := 0
		for {
			select {
			case <-checkTicker.C:
				checkCount++
				ac.logger.Printf("Starting check #%d", checkCount)

				result, err := ac.checkAvailability()
				if err != nil {
					ac.logger.Printf("Error in check #%d: %v", checkCount, err)
					continue
				}

				ac.logger.Printf("Check #%d completed. Appointments available: %v", checkCount, result.Available)

				// Notify users if availability changed
				if result.Available != ac.wasAvailable {
					ac.notifyUsers(result, false)
					ac.wasAvailable = result.Available
				}

			case <-autoCheckTicker.C:
				// Send updates to users who enabled auto-check
				result, err := ac.checkAvailability()
				if err != nil {
					ac.logger.Printf("Error in auto-check: %v", err)
					continue
				}
				ac.notifyUsers(result, true)
			}
		}
	}()

	// Set up updates configuration
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Start receiving updates
	updates := ac.bot.GetUpdatesChan(updateConfig)

	// Handle updates
	for update := range updates {
		ac.handleCommand(update)
	}
}

func (ac *AppointmentChecker) runOnce() {
	fmt.Println("Checking appointment availability...")
	result, err := ac.checkAvailability()
	if err != nil {
		fmt.Printf("❌ Error checking availability: %v\n", err)
		os.Exit(1)
		return
	}

	if result.Available {
		if result.AvailableSlot != nil {
			startTime, _ := time.Parse(time.RFC3339, result.AvailableSlot.StartDateTime.DateTime)
			fmt.Printf("🎉 Appointments are available!\nNext available slot: %s\nAvailable count: %d\n",
				startTime.Format("2006-01-02 15:04"), result.AvailableSlot.AvailableCount)
		} else {
			fmt.Println("🎉 Appointments are available!")
		}
		fmt.Println("\nBooking website: https://outlook.office365.com/owa/calendar/monetrapirkanmaarekrytointipalvelut@monetra.fi/bookings/")
	} else {
		fmt.Println("❌ No appointments available at the moment.")
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Command line flags
	botMode := flag.Bool("bot", false, "Run in bot mode (continuous checking)")
	flag.Parse()

	var checker *AppointmentChecker

	if *botMode {
		bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
		if err != nil {
			log.Fatal(err)
		}

		checker = NewAppointmentChecker(bot)
		checker.runBot()
	} else {
		checker = NewAppointmentChecker(nil)
		checker.runOnce()
	}
}
