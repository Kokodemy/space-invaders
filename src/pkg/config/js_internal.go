//go:build js && wasm

package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

const (
	originalWidth  = 760 // Original width of the drawable canvas area (px, after considering the padding and border of the surrounding containers)
	originalHeight = 570 // Original height of the drawable canvas area (px, after considering the padding and border of the surrounding containers)
)

const (
	animatedClickClass      = "animated-click"
	animatedClickEndClass   = "animated-click-end"
	audioIconId             = "audioIcon"
	audioIconMutedClass     = "fa-volume-mute"
	audioIconUnmutedClass   = "fa-volume-up"
	audioToggleBtnId        = "audioToggle"
	canvasId                = "gameCanvas"
	eventLogChannelButtonID = "eventLogChannelBtn"
	eventLogChannelID       = "eventLogChannel"
	goEnv                   = "go_env"
	infoLogChannelButtonID  = "infoLogChannelBtn"
	infoLogChannelID        = "infoLogChannel"
	messageBoxID            = "message"
	tabActiveClass          = "active"
	tabClass                = "tab"
	tabContentClass         = "tab-content"
	tabFlashClass           = "flashing"
	refreshButtonId         = "refreshButton"
)

var (
	audioCtx               = getAudioContext()
	audioPlayers           = make(map[string]audioPlayer)
	audioPlayersMutex      = sync.RWMutex{}
	audioTracks            = make(map[string][]byte)
	audioTracksMutex       = sync.RWMutex{}
	canvasObject           = document.Call("getElementById", canvasId)
	canvasObjectContext    = canvasObject.Call("getContext", "2d", MakeObject(map[string]any{"willReadFrequently": true}))
	console                = GlobalGet("console")
	document               = GlobalGet("document")
	eventLogChannel        = document.Call("getElementById", eventLogChannelID)
	eventLogChannelBtn     = document.Call("getElementById", eventLogChannelButtonID)
	fpsDiv                 = document.Call("getElementById", "fps")
	infoLogChannel         = document.Call("getElementById", infoLogChannelID)
	infoLogChannelBtn      = document.Call("getElementById", infoLogChannelButtonID)
	invisibleCanvas        = document.Call("createElement", "canvas")
	invisibleCtx           = invisibleCanvas.Call("getContext", "2d")
	invisibleCanvasScrollY = 0.0
	messageBox             = document.Call("getElementById", messageBoxID)
	lastLogSentTime        = time.Time{}
	scoreBoard             []score
	scoreBoardMutex        = sync.RWMutex{}
	window                 = GlobalGet("window")
	windowLocation         = window.Get("location")
)

// audioPlayer represents an audio player.
type audioPlayer struct {
	endedCallback js.Func
	source        js.Value
	startTime     float64
}

// dimensions represents the dimensions of the document.
type dimensions struct {
	BoxWidth, BoxHeight                  float64
	BoxLeft, BoxTop, BoxRight, BoxBottom float64
	OriginalWidth, OriginalHeight        float64
	ScaleWidth, ScaleHeight              float64
}

// logEvent represents a log event type.
type logEvent bool

// Channel returns the channel element.
func (e logEvent) Channel() js.Value {
	return map[logEvent]js.Value{true: eventLogChannel, false: infoLogChannel}[e]
}

// ChannelButton returns the channel button element.
func (e logEvent) ChannelButton() js.Value {
	return map[logEvent]js.Value{true: eventLogChannelBtn, false: infoLogChannelBtn}[e]
}

// score represents a score.
type score struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
}

// init is a function that initializes the game interface.
func init() {
	// Set up the game interface
	setupAudioInterface()
	setupCanvasInterface()
	setupMessageBoxInterface()
	setupRefreshInterface()
	setupScoreBoard()

	// Detach the watchdogs
	checkHealth(1)
	envCallback(1)
}

// checkHealth is a function that checks the health of the game.
// The health check is performed every 10 seconds.
// The health check is performed by fetching the health endpoint.
// If the health check fails, the next health check is scheduled with exponential backoff.
func checkHealth(exponentialBackoff float64) {
	delayInMs := 10_000 * time.Millisecond

	go func() {
		GlobalCall("fetch", "health", MakeObject(map[string]any{
			"method":  http.MethodGet,
			"headers": MakeObject(map[string]any{"Accept": "application/json"}),
		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			if response := p[0]; !response.Get("ok").Bool() {
				return response.Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
					return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
				}))
			}

			time.AfterFunc(delayInMs, func() { checkHealth(exponentialBackoff) })
			return nil

		})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) interface{} {
			LogError(fmt.Errorf("Error checking health: %s", p[0].String()))

			time.AfterFunc(delayInMs*time.Duration(exponentialBackoff), func() {
				checkHealth(exponentialBackoff * 2)
			})
			return nil

		}))
	}()
}

// envCallback is a function that fetches the environment variables.
func envCallback(exponentialBackoff float64) {
	delayInMs := 2_500 * time.Millisecond

	go func() {
		js.Global().Call("fetch", ".env", MakeObject(map[string]any{
			"method":  http.MethodGet,
			"headers": MakeObject(map[string]any{"Accept": "application/json"}),
		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			response := p[0]
			if !response.Get("ok").Bool() {
				return response.Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
					return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
				}))
			}

			return response.Call("json")

		})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
			envData := ConvertObjectToMap(p[0])
			prefix, _ := envData["_prefix"].(string)
			env := make(map[string]any)
			for key, value := range envData {
				if key != "_prefix" && len(prefix) > 0 && strings.HasPrefix(key, prefix) {
					env[key] = value
				}
			}

			if Config.Control.Debug.Get() {
				Log(fmt.Sprintf("Retrieved environment variables: %#v", env))
			}
			GlobalSet(goEnv, MakeObject(env))

			time.AfterFunc(delayInMs, func() {
				envCallback(exponentialBackoff)
			})

			return nil

		})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) interface{} {
			LogError(fmt.Errorf("Error getting env: %s", p[0].String()))

			time.AfterFunc(delayInMs*time.Duration(exponentialBackoff), func() {
				envCallback(exponentialBackoff * 2)
			})

			return nil
		}))
	}()
}

// getAudioContext is a function that returns the audio context.
func getAudioContext() js.Value {
	ctx := NewInstance("AudioContext")
	if !ctx.Truthy() {
		ctx = NewInstance("webkitAudioContext")
	}
	return ctx
}

// scoreBoardSortFunc is a function that sorts the scores.
// The scores are sorted in descending order of the score.
// If the scores have the same score, they are ordered in ascending order of their timestamps.
func scoreBoardSortFunc(i, j score) int {
	if i.Score == j.Score {
		if !i.UpdatedAt.IsZero() && !j.UpdatedAt.IsZero() {
			// Sort in ascending order
			return j.UpdatedAt.Compare(i.UpdatedAt)
		}

		// Sort in ascending order
		return j.CreatedAt.Compare(i.CreatedAt)
	}

	// Sort in descending order
	return j.Score - i.Score
}

// setupAudioInterface is a function that sets up the audio interface.
// The audio interface includes the audio icon and the audio toggle button.
// The audio icon is updated based on the audio state.
// The audio toggle button toggles the audio state.
// The audio toggle button is updated based on the audio state.
func setupAudioInterface() {
	audioIcon := document.Call("getElementById", audioIconId)

	if *Config.Control.AudioEnabled {
		audioIcon.Get("classList").Call("remove", audioIconMutedClass)
		audioIcon.Get("classList").Call("add", audioIconUnmutedClass)
	} else {
		audioIcon.Get("classList").Call("remove", audioIconUnmutedClass)
		audioIcon.Get("classList").Call("add", audioIconMutedClass)
	}

	audioToggle := func() {
		*Config.Control.AudioEnabled = !*Config.Control.AudioEnabled

		if *Config.Control.AudioEnabled {
			audioIcon.Get("classList").Call("remove", audioIconMutedClass)
			audioIcon.Get("classList").Call("add", audioIconUnmutedClass)

			go PlayAudio("theme_heroic.wav", true)
		} else {
			audioIcon.Get("classList").Call("remove", audioIconUnmutedClass)
			audioIcon.Get("classList").Call("add", audioIconMutedClass)

			go StopAudioSources(func(string) bool { return true })
		}
	}

	GlobalSet("toggleAudioClick", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		audioToggle()
		return nil
	}))

	GlobalSet("toggleAudioTouchEnd", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		audioToggle()
		return nil
	}))

	audioToggleBtn := document.Call("getElementById", audioToggleBtnId)
	audioToggleBtn.Call("addEventListener", "click", GlobalGet("toggleAudioClick"))
	audioToggleBtn.Call("addEventListener", "touchend", GlobalGet("toggleAudioTouchEnd"))
}

// setupCanvasInterface is a function that sets up the canvas interface.
// The canvas interface includes the invisible canvas and the resize event listener.
// The invisible canvas is used to draw the background of the visible canvas.
// The resize event listener is used to redraw the document when the window is resized.
// The draw function is called when the document is resized.
func setupCanvasInterface() {
	invisibleCanvas.Set("width", canvasObject.Get("width").Float())
	invisibleCanvas.Set("height", canvasObject.Get("height").Float())

	GlobalSet("resize", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		if GlobalGet("drawFunc").Truthy() {
			GlobalGet("drawFunc").Invoke()
		}
		return nil
	}))

	window.Call("addEventListener", "resize", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		GlobalCall("requestAnimationFrame", GlobalGet("resize"))
		return nil
	}))

	GlobalCall("requestAnimationFrame", GlobalGet("resize"))
}

// setupMessageBoxInterface is a function that sets up the message box interface.
// The message box is scrollable only if the content inside the #message element can scroll.
// The touch events are prevented from propagating to the body when the message box is touched.
func setupMessageBoxInterface() {
	openTab := func(event js.Value, channelId string) {
		// Hide all tab contents
		tabContents := document.Call("getElementsByClassName", tabContentClass)
		for i := 0; i < tabContents.Length(); i++ {
			tabContents.Index(i).Get("classList").Call("remove", tabActiveClass)
		}

		// Remove active class from all tabs
		tabs := document.Call("getElementsByClassName", tabClass)
		for i := 0; i < tabs.Length(); i++ {
			tabs.Index(i).Get("classList").Call("remove", tabActiveClass)
		}

		// Show the current tab content and add active class to the clicked tab
		tabContent := document.Call("getElementById", channelId)
		tabContent.Get("classList").Call("add", tabActiveClass)
		event.Get("currentTarget").Get("classList").Call("add", "active")

		// Scroll to the bottom of the tab content
		tabContent.Set("scrollTop", tabContent.Get("scrollHeight"))
	}

	eventLogChannelBtn.Call("addEventListener", "click", js.FuncOf(func(_ js.Value, p []js.Value) any {
		openTab(p[0], eventLogChannelID)
		return nil
	}))

	infoLogChannelBtn.Call("addEventListener", "click", js.FuncOf(func(_ js.Value, p []js.Value) any {
		openTab(p[0], infoLogChannelID)
		return nil
	}))

	if !IsTouchDevice() {
		return
	}

	// Apply the touch event listeners to the message box
	messageBox.Call("addEventListener", "touchstart", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("stopPropagation") // Stop the touch event from propagating to the body
		return nil
	}))

	messageBox.Call("addEventListener", "touchmove", js.FuncOf(func(_ js.Value, p []js.Value) any {
		// Allow touch move events only if the content inside the #message element can scroll
		if messageBox.Get("scrollHeight").Float() > messageBox.Get("clientHeight").Float() {
			p[0].Call("stopPropagation") // Prevent body scroll when inside the messageBox
		}
		return nil
	}))
}

// setupRefreshInterface is a function that sets up the refresh interface.
// The refresh interface includes the refresh button.
// The refresh button is animated when clicked or touched.
// The document is reloaded when the refresh button is clicked or touched.
func setupRefreshInterface() {
	refreshButton := document.Call("getElementById", refreshButtonId)
	GlobalSet("animateRefreshButton", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return NewInstance("Promise", js.FuncOf(func(_ js.Value, p []js.Value) any {
			refreshButton.Get("classList").Call("add", animatedClickClass)
			resolve := p[0]

			refreshButton.Call("addEventListener", "transitionend", js.FuncOf(func(_ js.Value, _ []js.Value) any {
				refreshButton.Get("classList").Call("remove", animatedClickClass)
				refreshButton.Get("classList").Call("add", animatedClickEndClass)
				resolve.Invoke()

				return nil
			}), MakeObject(map[string]any{"once": true}))
			return nil
		}))
	}))

	then := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		windowLocation.Call("reload")
		return nil
	})

	GlobalSet("refreshButtonClick", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		GlobalGet("animateRefreshButton").Invoke().Call("then", then)
		return nil
	}))

	GlobalSet("refreshButtonTouchEnd", js.FuncOf(func(_ js.Value, p []js.Value) any {
		p[0].Call("preventDefault")
		GlobalGet("animateRefreshButton").Invoke().Call("then", then)
		return nil
	}))

	refreshButton.Call("addEventListener", "click", GlobalGet("refreshButtonClick"))
	refreshButton.Call("addEventListener", "touchend", GlobalGet("refreshButtonTouchEnd"))
}

// setupScoreBoard is a function that sets up the score board.
func setupScoreBoard() {
	scoreBoardMutex.Lock()

	GlobalCall("fetch", "scores.db", MakeObject(map[string]any{
		"method":  http.MethodGet,
		"headers": MakeObject(map[string]any{"Content-Type": "application/json"}),
	})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
		if !p[0].Get("ok").Bool() {
			return p[0].Call("text").Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
				return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(p[0].String()))
			}))
		}

		return p[0].Call("text")

	})).Call("then", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		if err := json.Unmarshal([]byte(strings.TrimPrefix(p[0].String(), "while(1);")), &scoreBoard); err != nil {
			return GlobalGet("Promise").Call("reject", GlobalGet("Error").New(err.Error()))
		}

		slices.SortStableFunc(scoreBoard, scoreBoardSortFunc)
		if Config.Control.Debug.Get() {
			Log(fmt.Sprintf("Fetched score board: %#v", scoreBoard))
		}

		return nil

	})).Call("catch", js.FuncOf(func(_ js.Value, p []js.Value) any {
		defer scoreBoardMutex.Unlock()

		LogError(fmt.Errorf("failed to load score board: %s", p[0].String()))
		return nil

	}))
}
