package spaceship

import (
	"testing"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// TestEarlyLevelsAreNotFree pins that a level always costs more than the
// experience a single weakest kill is worth. The exponential term alone rounded
// to 1 for the first fourteen levels, so one kill of a level one enemy — worth
// DefaultPenalty, three — handed out three levels at once.
func TestEarlyLevelsAreNotFree(t *testing.T) {
	weakestKill := config.Config.Enemy.DefaultPenalty

	for progress := 1; progress <= 30; progress++ {
		lvl := SpaceshipLevel{Progress: progress}

		if required := lvl.GetRequiredExperience(); required <= weakestKill/2 {
			t.Errorf("progress %d: a level costs %d experience, less than half of a single kill (%d)",
				progress, required, weakestKill)
		}
	}
}

// TestRequiredExperienceIsMonotonic pins that the curve never dips, which would
// let a level become cheaper than the one before it.
func TestRequiredExperienceIsMonotonic(t *testing.T) {
	previous := 0
	for progress := 1; progress <= 500; progress++ {
		required := SpaceshipLevel{Progress: progress}.GetRequiredExperience()
		if required < previous {
			t.Fatalf("progress %d: required experience fell from %d to %d", progress, previous, required)
		}

		previous = required
	}
}

// TestPenaltyCannotWipeARun pins the cap on a single collision. The heavy end of
// the roster carries penalties in the hundreds, so one touch from an Overlord
// used to take a long run's whole progress and the shield's capacity with it.
func TestPenaltyCannotWipeARun(t *testing.T) {
	const progress = 100

	spaceship := Embark("test")
	for spaceship.Level.Progress < progress {
		spaceship.Level.Up()
	}

	// Spend the shield first, so that the levels rather than the charges are
	// what the penalty has to come out of.
	spaceship.Level.Shield.Charge = 0

	spaceship.Penalize(config.Config.Enemy.Overlord.Penalty)

	if spaceship.Level.Progress == 0 {
		t.Fatal("a single collision destroyed a spaceship at progress 100")
	}

	if lost := progress - spaceship.Level.Progress; lost > progress/4 {
		t.Errorf("a single collision cost %d of %d levels, want at most a quarter", lost, progress)
	}
}

// TestEnemyCeilingFollowsTheSpaceship pins that the roster escalates with the
// player's progress rather than with time survived.
func TestEnemyCeilingFollowsTheSpaceship(t *testing.T) {
	if got := enemy.MaximumType(0); got != enemy.Berserker {
		t.Errorf("a fresh run may field %s, want Berserker", got)
	}

	if got := enemy.MaximumType(1_000); got != enemy.Overlord {
		t.Errorf("a long run may field %s, want Overlord", got)
	}

	previous := enemy.MaximumType(0)
	for progress := 0; progress <= 1_000; progress += 5 {
		if got := enemy.MaximumType(progress); got < previous {
			t.Fatalf("progress %d: ceiling fell from %s to %s", progress, previous, got)
		} else {
			previous = got
		}
	}
}
