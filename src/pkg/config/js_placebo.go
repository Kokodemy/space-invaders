//go:build !js || !wasm

package config

import (
	"log"
	"os"
	"time"
)

type dimensions struct {
	BoxWidth, BoxHeight                  float64
	BoxLeft, BoxTop, BoxRight, BoxBottom float64
	OriginalWidth, OriginalHeight        float64
	ScaleWidth, ScaleHeight              float64
}

type score struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func AddEventListener(event string, listener any)         {}
func AddEventListenerToCanvas(event string, listener any) {}

func CanvasBoundingBox() dimensions {
	return dimensions{BoxWidth: 800, BoxHeight: 600, BoxRight: 800, BoxBottom: 600}
}

func ClearBackground()                                                                {}
func ClearCanvas()                                                                    {}
func ConvertArrayToSlice(array any) []any                                             { return nil }
func ConvertObjectToMap(obj any) map[string]any                                       { return nil }
func DrawAnomalyBlackHole(coords [2]float64, radius float64)                          {}
func DrawAnomalySupernova(coords [2]float64, radius float64)                          {}
func DrawBackground(speed float64)                                                    {}
func DrawLine(start, end [2]float64, color string, thickness float64)                 {}
func DrawPlanetEarth(coords [2]float64, radius float64)                               {}
func DrawPlanetJupiter(coords [2]float64, radius float64)                             {}
func DrawPlanetMars(coords [2]float64, radius float64)                                {}
func DrawPlanetMercury(coords [2]float64, radius float64)                             {}
func DrawPlanetNeptune(coords [2]float64, radius float64)                             {}
func DrawPlanetPluto(coords [2]float64, radius float64)                               {}
func DrawPlanetSaturn(coords [2]float64, radius float64)                              {}
func DrawPlanetUranus(coords [2]float64, radius float64)                              {}
func DrawPlanetVenus(coords [2]float64, radius float64)                               {}
func DrawRect(coords [2]float64, size [2]float64, color string, cornerRadius float64) {}

func DrawSpaceship(coors [2]float64, size [2]float64, faceUp bool, color, label string, statusValues []float64, statusColors []string) {
}

func DrawStar(coords [2]float64, spikes int, radius, innerRadius float64, color string, brightness float64) {
}

func DrawSun(coords [2]float64, radius float64) {}
func Getenv(key string) string                  { return os.Getenv(key) }
func GetScores(top int) (scores []score)        { return }
func GlobalCall(name string, args ...any) any   { return nil }
func GlobalGet(key string) any                  { return nil }
func GlobalSet(key string, value any)           {}
func IsPlaying(name string) bool                { return false }
func IsTouchDevice() bool                       { return false }
func LoadAudio(url string) ([]byte, error)      { return nil, nil }
func Log(msg string)                            { log.Println(msg) }

func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func MakeObject(m map[string]any) any                                            { return m }
func NewInstance(typ string, args ...any) any                                    { return nil }
func PlayAudio(name string, loop bool)                                           {}
func SaveScores()                                                                {}
func SendMessage(msg string, reset, event bool)                                  { log.Println(msg) }
func SendMessageThrottled(msg string, reset, event bool, cooldown time.Duration) { log.Println(msg) }
func Setenv(key, value string)                                                   { _ = os.Setenv(key, value) }
func SetScore(name string, score int) (rank int)                                 { return }
func StopAudio(name string)                                                      {}
func StopAudioSources(selector func(name string) bool)                           {}

func ThrowError(err error) {
	if err != nil {
		panic(err)
	}
}

func Unsetenv(key string) { _ = os.Unsetenv(key) }

func UpdateFPS(fps float64) {}
