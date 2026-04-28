package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/starfall-warsong/sws/internal/model"
)

type CombatEngine struct {
	State *model.CombatState
	rng   *rand.Rand
}

func NewCombatEngine(combatID int64) *CombatEngine {
	return &CombatEngine{
		State: &model.CombatState{
			CombatID: combatID,
			Tick:     0,
			Status:   "active",
		},
		rng: rand.New(rand.NewSource(combatID)),
	}
}

func (e *CombatEngine) AddParticipant(p model.CombatParticipant) {
	e.State.Participants = append(e.State.Participants, p)
}

func (e *CombatEngine) ProcessTick() []string {
	e.State.Tick++
	var logs []string

	logs = append(logs, fmt.Sprintf("═══ Tick %d ═══", e.State.Tick))

	// Auto-target for participants without a target
	for i := range e.State.Participants {
		p := &e.State.Participants[i]
		if p.IsDestroyed {
			continue
		}
		if p.TargetID == nil || e.isDestroyed(*p.TargetID) {
			target := e.findTarget(p)
			if target != nil {
				p.TargetID = &target.ID
			}
		}
	}

	// Process movement: NPCs approach their optimal range
	for i := range e.State.Participants {
		p := &e.State.Participants[i]
		if p.IsDestroyed || p.TargetID == nil {
			continue
		}
		target := e.getParticipant(*p.TargetID)
		if target == nil {
			continue
		}
		// Move towards optimal range
		movePerTick := p.Speed * 3 // 3 seconds per tick
		if p.Distance > p.OptimalRange+2000 {
			p.Distance -= movePerTick
			if p.Distance < p.OptimalRange {
				p.Distance = p.OptimalRange
			}
		} else if p.Distance < p.OptimalRange-2000 && p.Type == "npc" {
			p.Distance += movePerTick / 2
		}
		// Sync distance for target (simplified: both at same distance)
		target.Distance = p.Distance
	}

	// Process damage from each participant
	for i := range e.State.Participants {
		attacker := &e.State.Participants[i]
		if attacker.IsDestroyed || attacker.TargetID == nil {
			continue
		}

		// 射速检查：rate_of_fire=N 表示每N个Tick开一次火
		rof := attacker.RateOfFire
		if rof <= 0 {
			rof = 1
		}
		if e.State.Tick%rof != 0 {
			continue // 这个Tick不开火
		}

		target := e.getParticipant(*attacker.TargetID)
		if target == nil || target.IsDestroyed {
			continue
		}

		// Calculate hit chance
		hitChance := e.calculateHitChance(attacker, target)
		if e.rng.Float64() > hitChance {
			logs = append(logs, fmt.Sprintf("  %s 对 %s 的攻击未命中", attacker.Name, target.Name))
			continue
		}

		// Range falloff
		rangeMod := e.calculateRangeMod(attacker, target)
		rawDamage := int(float64(attacker.DamagePerTick) * rangeMod)

		// Apply damage to target
		actualDamage, hitLayer := e.applyDamage(target, rawDamage, attacker.DamageType)

		logs = append(logs, fmt.Sprintf("  %s 的%s命中 %s 的%s，造成 %d 伤害",
			attacker.Name, damageTypeName(attacker.DamageType), target.Name, hitLayer, actualDamage))

		if target.StructureCurrent <= 0 {
			target.IsDestroyed = true
			logs = append(logs, fmt.Sprintf("  ★ %s 被击毁！", target.Name))
		}
	}

	// Shield recharge
	for i := range e.State.Participants {
		p := &e.State.Participants[i]
		if p.IsDestroyed {
			continue
		}
		if p.ShieldCurrent < p.ShieldMax {
			p.ShieldCurrent += p.ShieldRecharge
			if p.ShieldCurrent > p.ShieldMax {
				p.ShieldCurrent = p.ShieldMax
			}
		}
	}

	// Check combat end
	teamAAlive := false
	teamBAlive := false
	for _, p := range e.State.Participants {
		if p.IsDestroyed {
			continue
		}
		if p.Team == "a" {
			teamAAlive = true
		} else {
			teamBAlive = true
		}
	}

	if !teamAAlive || !teamBAlive {
		e.State.Status = "finished"
		if teamAAlive {
			logs = append(logs, "═══ 战斗结束: A方胜利 ═══")
		} else if teamBAlive {
			logs = append(logs, "═══ 战斗结束: B方胜利 ═══")
		} else {
			logs = append(logs, "═══ 战斗结束: 双方全灭 ═══")
		}
	}

	e.State.Logs = logs
	return logs
}

func (e *CombatEngine) calculateHitChance(attacker, target *model.CombatParticipant) float64 {
	sigRatio := float64(target.Signature) / 100.0
	baseHit := 0.7 * sigRatio
	if baseHit > 0.95 {
		baseHit = 0.95
	}
	if baseHit < 0.1 {
		baseHit = 0.1
	}
	return baseHit
}

func (e *CombatEngine) calculateRangeMod(attacker, target *model.CombatParticipant) float64 {
	dist := target.Distance
	optimal := attacker.OptimalRange
	if dist <= optimal {
		return 1.0
	}
	falloff := float64(dist-optimal) / float64(optimal)
	mod := math.Exp(-falloff * falloff)
	if mod < 0.1 {
		return 0.1
	}
	return mod
}

func (e *CombatEngine) applyDamage(target *model.CombatParticipant, damage int, dmgType model.DamageType) (int, string) {
	totalApplied := 0
	remaining := damage

	// Shield first — 使用该船的护盾抗性
	if target.ShieldCurrent > 0 && remaining > 0 {
		resist := getResistFromProfile(target.ShieldResist, dmgType)
		shieldDmg := int(float64(remaining) * (1.0 - resist))
		if target.ShieldCurrent >= shieldDmg {
			target.ShieldCurrent -= shieldDmg
			totalApplied += shieldDmg
			return totalApplied, fmt.Sprintf("护盾(抗%.0f%%)", resist*100)
		}
		remaining = remaining - int(float64(target.ShieldCurrent)/(1.0-resist))
		totalApplied += target.ShieldCurrent
		target.ShieldCurrent = 0
	}

	// Armor — 使用该船的装甲抗性
	if target.ArmorCurrent > 0 && remaining > 0 {
		resist := getResistFromProfile(target.ArmorResist, dmgType)
		armorDmg := int(float64(remaining) * (1.0 - resist))
		if target.ArmorCurrent >= armorDmg {
			target.ArmorCurrent -= armorDmg
			totalApplied += armorDmg
			return totalApplied, fmt.Sprintf("装甲(抗%.0f%%)", resist*100)
		}
		remaining = remaining - int(float64(target.ArmorCurrent)/(1.0-resist))
		totalApplied += target.ArmorCurrent
		target.ArmorCurrent = 0
	}

	// Structure — 统一5%抗性
	if remaining > 0 {
		structDmg := int(float64(remaining) * 0.95)
		target.StructureCurrent -= structDmg
		totalApplied += structDmg
		if target.StructureCurrent < 0 {
			target.StructureCurrent = 0
		}
	}

	return totalApplied, "结构(抗5%)"
}

func getResistFromProfile(profile model.ResistProfile, dmgType model.DamageType) float64 {
	switch dmgType {
	case model.DamageKinetic:
		return profile.Kinetic
	case model.DamageThermal:
		return profile.Thermal
	case model.DamageEM:
		return profile.EM
	case model.DamageExplosive:
		return profile.Explosive
	default:
		return 0.1
	}
}

func (e *CombatEngine) findTarget(p *model.CombatParticipant) *model.CombatParticipant {
	for i := range e.State.Participants {
		t := &e.State.Participants[i]
		if t.Team != p.Team && !t.IsDestroyed {
			return t
		}
	}
	return nil
}

func (e *CombatEngine) getParticipant(id int64) *model.CombatParticipant {
	for i := range e.State.Participants {
		if e.State.Participants[i].ID == id {
			return &e.State.Participants[i]
		}
	}
	return nil
}

func (e *CombatEngine) isDestroyed(id int64) bool {
	p := e.getParticipant(id)
	return p == nil || p.IsDestroyed
}

func damageTypeName(dt model.DamageType) string {
	switch dt {
	case model.DamageKinetic:
		return "动能弹"
	case model.DamageThermal:
		return "热能束"
	case model.DamageEM:
		return "电磁脉冲"
	case model.DamageExplosive:
		return "爆炸弹"
	default:
		return "攻击"
	}
}
