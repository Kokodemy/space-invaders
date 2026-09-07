package planet

import (
	"testing"

	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// TestApplyGravityScalesWithTheFrame pins that the pull is a displacement per
// unit of simulated time rather than per frame. Unscaled it was applied once per
// frame however long the frame lasted, so every planet pulled about 2.4 times
// harder on a 144 Hz display than on a 60 Hz one.
func TestApplyGravityScalesWithTheFrame(t *testing.T) {
	const mass = 600

	// Far enough out that the clamp against the remaining distance cannot engage
	// and hide the scaling.
	origin := numeric.Locate(400, 400)

	for _, kind := range []PlanetType{Earth, Sun, BlackHole, Supernova} {
		single := &Planet{Position: numeric.Locate(400, 100), Radius: 50, Type: kind}
		double := &Planet{Position: numeric.Locate(400, 100), Radius: 50, Type: kind}

		moved := single.ApplyGravity(origin, mass, false, false, 1).Sub(origin)
		twice := double.ApplyGravity(origin, mass, false, false, 2).Sub(origin)

		if moved.IsZero() {
			t.Fatalf("%s: no pull at all, the test proves nothing", kind)
		}

		if !numeric.Equal(twice, moved.Mul(2), 1e-9) {
			t.Errorf("%s: a double-length frame moved %s, want %s", kind, twice, moved.Mul(2))
		}
	}
}

// TestApplyGravityNeverOvershootsTheCentre pins that the clamp still holds once
// the frame scale is applied, so that a long frame cannot carry an object past
// the body pulling it.
func TestApplyGravityNeverOvershootsTheCentre(t *testing.T) {
	hole := &Planet{Position: numeric.Locate(400, 400), Radius: 100, Type: BlackHole}
	origin := numeric.Locate(400, 402) // Practically on top of it.

	// The largest scale a frame is allowed to carry.
	if moved := hole.ApplyGravity(origin, 1e6, false, false, 4); moved.Distance(hole.Position) > 1e-9 {
		t.Errorf("gravity carried the object to %s, past the centre at %s", moved, hole.Position)
	}
}
