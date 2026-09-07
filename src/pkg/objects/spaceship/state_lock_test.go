package spaceship

import (
	"testing"
	"time"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

// TestFreezeCannotBeChainedIndefinitely pins the bound on a disabling state.
// A frozen spaceship can neither move nor shoot, so every further freezer that
// reached it used to restart the timer on a player with no way to break the
// chain. Prolonging is now measured from when the state was entered, so the
// episode ends whatever the enemies do.
func TestFreezeCannotBeChainedIndefinitely(t *testing.T) {
	if testing.Short() {
		// The bound is wall-clock, so proving it costs the length of one capped
		// episode.
		t.Skip("waits out a full freeze episode")
	}

	spaceship := Embark("test")
	spaceship.ChangeState(Frozen)

	maximum := time.Duration(float64(config.Config.Spaceship.FreezeDuration) *
		config.Config.Spaceship.MaximumStateDurationFactor)

	// Stand in for a stream of freezers arriving every frame for far longer than
	// the episode may last.
	started := time.Now()
	for time.Since(started) < maximum+time.Second {
		spaceship.ChangeState(Frozen) // Another freezer lands on a helpless ship.
		spaceship.UpdateState(1)

		if spaceship.State() != Frozen {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	if spaceship.State() == Frozen {
		t.Fatalf("still frozen after %s, cap is %s", time.Since(started), maximum)
	}

	if held := time.Since(started); held > maximum+250*time.Millisecond {
		t.Errorf("freeze lasted %s, want at most about %s", held, maximum)
	}
}

// TestDisablingStateGracePeriod pins that a freeze cannot be reapplied the
// instant the previous one lifts, which is what gives the player a chance to fly
// clear before the next freezer arrives.
func TestDisablingStateGracePeriod(t *testing.T) {
	spaceship := Embark("test")
	spaceship.ChangeState(Frozen)
	spaceship.ResetState()

	spaceship.ChangeState(Frozen)
	if spaceship.State() == Frozen {
		t.Error("spaceship was re-frozen inside the grace period")
	}

	// A state that does not take control away is unaffected by the grace period.
	spaceship.ChangeState(Damaged)
	if spaceship.State() != Damaged {
		t.Errorf("state is %s, want Damaged", spaceship.State())
	}
}

// TestGravitationalMassIsNotInflatedByTheBoost pins that the pull of a planet
// does not depend on the state the spaceship is in. The boost scales the hull by
// half again, so taking the gravitational mass from the live area made every
// planet pull 2.25 times harder for the duration of the reward.
func TestGravitationalMassIsNotInflatedByTheBoost(t *testing.T) {
	spaceship := Embark("test")

	neutralArea, neutralMass := spaceship.Area(), spaceship.GravitationalMass()

	spaceship.ChangeState(Boosted)
	for i := 0; i < 1_000; i++ { // Let the size transition run to completion.
		spaceship.Geometry.Interpolate(1)
	}

	// Guard against the test passing because the hull never grew.
	if boostedArea := spaceship.Area(); boostedArea <= neutralArea {
		t.Fatalf("hull did not grow with the boost: %v -> %v", neutralArea, boostedArea)
	}

	if got := spaceship.GravitationalMass(); !numeric.Equal(got/neutralMass, 1, 1e-6) {
		t.Errorf("gravitational mass changed with the boost: %v -> %v", neutralMass, got)
	}
}
