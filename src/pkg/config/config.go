package config

import (
	_ "embed"
	"fmt"
	"reflect"
	"time"

	"gopkg.in/ini.v1"
)

//go:embed config.ini
var configFile []byte

// Config is the configuration of the game.
var Config config = func() (config config) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys:          true,
		DebugFunc:                 Log,
		AllowNestedValues:         true,
		Insensitive:               true,
		SkipUnrecognizableLines:   true,
		UnescapeValueDoubleQuotes: true,
	}, configFile)
	ThrowError(err)

	// Set the block mode to false to speed up the loading.
	cfg.BlockMode = false
	ThrowError(cfg.MapTo(&config))

	// Sanitize the configuration.
	ThrowError(config.sanitize())
	return
}()

// config represents the configuration of the game.
type config struct {
	Bullet struct {
		CriticalHitChance    float64
		CriticalHitFactor    int
		Height               float64
		InitialDamage        int
		ModifierProgressStep int
		Speed                float64
		WeightFactor         float64
		Width                float64
	}

	Control struct {
		AudioEnabled                      *bool
		BackgroundAnimationEnabled        *bool
		CollisionDetectionVersion         EnvVariable[int]
		CriticalFramesPerSecondRate       float64
		Debug                             EnvVariable[bool]
		DesiredFramesPerSecondRate        float64
		DrawEnemyHitpointBars             EnvVariable[bool]
		DrawObjectLabels                  EnvVariable[bool]
		DrawSpaceshipDiscoveryProgressBar EnvVariable[bool]
		DrawSpaceshipExperienceBar        EnvVariable[bool]
		DrawSpaceshipShield               EnvVariable[bool]
		GodMode                           EnvVariable[bool]
		SuspensionFrames                  int
	}

	Enemy struct {
		AccelerationProgress      float64
		Count                     int
		CountProgressStep         int
		BerserkLikeliness         float64
		BerserkLikelinessProgress float64
		DefaultPenalty            int
		DefenseProgress           int
		Height                    float64
		HitpointProgress          int
		InitialDefense            int
		InitialHitpoints          int
		InitialSpeed              float64
		MaximumCount              int
		MaximumSpeed              float64
		Regenerate                *bool
		SpecialtyLikeliness       float64
		Width                     float64

		Annihilator struct {
			DefenseBoost    int
			HitpointsBoost  int
			Penalty         int
			SizeFactorBoost float64
			SpeedModifier   float64
			YetAgainFactor  int
		} `ini:"Enemy.Annihilator"`

		Berserker struct {
			DefenseBoost    int
			HitpointsBoost  int
			Penalty         int
			SizeFactorBoost float64
			SpeedModifier   float64
		} `ini:"Enemy.Berserker"`

		Freezer struct {
			Penalty int
		} `ini:"Enemy.Freezer"`
	}

	MessageBox struct {
		BufferSize    int
		LogThrottling time.Duration

		Messages struct {
			AllPlanetsDiscovered         TemplateString
			EnemyDestroyed               TemplateString
			EnemyHit                     TemplateString
			ExplainInterface             TemplateString
			GamePaused                   TemplateString
			GameStarted                  TemplateString
			GameOver                     TemplateString
			Greeting                     TemplateString
			HowToRestart                 TemplateString
			PerformanceDropped           TemplateString
			PerformanceImproved          TemplateString
			PlanetDiscovered             TemplateString
			PlanetImpactsSystem          TemplateString
			Prompt                       TemplateString
			ScoreBoardUpdated            TemplateString
			SpaceshipDowngradedByEnemy   TemplateString
			SpaceshipFrozen              TemplateString
			SpaceshipStillFrozen         TemplateString
			SpaceshipUpgradedByEnemyKill TemplateString
			SpaceshipUpgradedByGoodie    TemplateString
			WaitForScoreBoardUpdate      TemplateString
		} `ini:"MessageBox.Messages"`
	}

	Planet struct {
		DiscoveryCooldown    time.Duration
		DiscoveryProbability float64
		MaximumRadius        float64
		MinimumRadius        float64
		SpeedRatio           float64

		Impact struct {
			DefaultGravityStrength float64

			Mercury struct {
				BerserkLikelinessAmplifier float64
				Description                TemplateString
			} `ini:"Planet.Impact.Mercury"`

			Venus struct {
				Description               TemplateString
				GoodieLikelinessAmplifier float64
				SpaceshipDeceleration     float64
			} `ini:"Planet.Impact.Venus"`

			Earth struct {
				Description               TemplateString
				GoodieLikelinessAmplifier float64
				SpaceshipDeceleration     float64
			} `ini:"Planet.Impact.Earth"`

			Mars struct {
				BerserkLikelinessAmplifier float64
				Description                TemplateString
			} `ini:"Planet.Impact.Mars"`

			Jupiter struct {
				Description             TemplateString
				EnemyDefenseAmplifier   int
				EnemyHitpointsAmplifier int
			} `ini:"Planet.Impact.Jupiter"`

			Saturn struct {
				Description             TemplateString
				EnemyDefenseAmplifier   int
				EnemyHitpointsAmplifier int
			} `ini:"Planet.Impact.Saturn"`

			Uranus struct {
				Description                TemplateString
				FreezerLikelinessAmplifier float64
			} `ini:"Planet.Impact.Uranus"`

			Neptune struct {
				Description                TemplateString
				FreezerLikelinessAmplifier float64
			} `ini:"Planet.Impact.Neptune"`

			Pluto struct {
				BerserkLikelinessAmplifier float64
				Description                TemplateString
				FreezerLikelinessAmplifier float64
			} `ini:"Planet.Impact.Pluto"`

			Sun struct {
				Description     TemplateString
				GravityStrength float64
			} `ini:"Planet.Impact.Sun"`

			BlackHole struct {
				Description     TemplateString
				GravityStrength float64
			} `ini:"Planet.Impact.BlackHole"`

			Supernova struct {
				Description     TemplateString
				GravityStrength float64
			} `ini:"Planet.Impact.Supernova"`
		} `ini:"Planet.Impact"`
	}

	Spaceship struct {
		Acceleration           float64
		AdmiralDamageAmplifier int
		BoostDuration          time.Duration
		BoostScaleSizeFactor   float64
		CannonProgress         int
		Cooldown               time.Duration
		DamageDuration         time.Duration
		ExperienceScaler       float64
		FreezeDuration         time.Duration
		Height                 float64
		MaximumCannons         int
		MaximumLabelLength     int
		MaximumSpeed           float64
		ShieldChargeDuration   time.Duration
		Width                  float64
	}

	Star struct {
		Brightness         float64
		Count              int
		MinimumInnerRadius float64
		MinimumRadius      float64
		MinimumSpikes      float64
		MaximumInnerRadius float64
		MaximumRadius      float64
		MaximumSpikes      float64
		SpeedRatio         float64
	}
}

// Sanitize sanitizes the configuration.
// It calls the Sanitize method of each field that has one.
// The Sanitize method should have the following signature:
// func (Type) Sanitize() Type
func (cfg *config) sanitize() error {
	const methodName = "Sanitize"

	var sanitize func(reflect.Value) error
	sanitize = func(v reflect.Value) error {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Struct {
			for i := 0; i < v.NumField(); i++ {
				if err := sanitize(v.Field(i)); err != nil {
					return err
				}
			}
		}

		if !v.CanAddr() {
			return fmt.Errorf("cannot take the address of %v", v.Type())
		}

		method, ok := v.Type().MethodByName(methodName)
		if !ok {
			return nil
		}

		if method.Type.NumIn() != 1 || method.Type.NumOut() != 1 || !method.Type.Out(0).AssignableTo(v.Type()) {
			return fmt.Errorf("invalid signature for %[1]s method: %[2]v, expected: func (%[3]v) %[1]s() %[3]v",
				methodName, method.Type, v.Type())
		}

		v.Set(v.Addr().MethodByName(methodName).Call(nil)[0])
		return nil
	}

	return sanitize(reflect.ValueOf(cfg).Elem())
}
