package spaceship

import (
	"testing"

	"github.com/sarumaj/edu-space-invaders/src/pkg/config"
	"github.com/sarumaj/edu-space-invaders/src/pkg/objects/enemy"
)

// TestOpeningLevelsStayCheap pins that the first levels cost about what they
// always cost. Charging a baseline for them instead made the opening two to four
// times slower, and slower levelling means weaker bullets, which means enemies
// live longer and more of them reach the bottom — so the opening got harder than
// the arithmetic alone suggests. The fast opening ramp is the design.
func TestOpeningLevelsStayCheap(t *testing.T) {
	weakestKill := config.Config.Enemy.DefaultPenalty

	for progress := 1; progress <= 10; progress++ {
		if required := (SpaceshipLevel{Progress: progress}).GetRequiredExperience(); required > weakestKill {
			t.Errorf("progress %d: a level costs %d experience, more than a single kill is worth (%d)",
				progress, required, weakestKill)
		}
	}
}

// TestLateLevelsStayReachable pins the absence of the wall at the other end. The
// original curve reached 4160 experience for a single level at progress 300 —
// well over a thousand kills for one level — against a gain that grows roughly
// linearly, so progress stopped rather than slowed.
func TestLateLevelsStayReachable(t *testing.T) {
	// A level should never be worth more than about a hundred of the weakest kills.
	budget := 100 * config.Config.Enemy.DefaultPenalty

	if required := (SpaceshipLevel{Progress: 300}).GetRequiredExperience(); required > budget {
		t.Errorf("a level at progress 300 costs %d experience, more than the %d budget", required, budget)
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
