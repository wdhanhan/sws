package engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/starfall-warsong/sws/internal/model"
)

type StarmapGenerator struct {
	rng *rand.Rand
}

func NewStarmapGenerator(masterSeed int64) *StarmapGenerator {
	return &StarmapGenerator{rng: rand.New(rand.NewSource(masterSeed))}
}

type GeneratedMap struct {
	Systems  []model.StarSystem
	Gates    []model.Stargate
	Planets  []model.Planet
	Belts    []model.AsteroidBelt
}

func (g *StarmapGenerator) Generate(systemCount int) *GeneratedMap {
	result := &GeneratedMap{}

	armDistribution := map[model.ArmID]float64{
		model.ArmFire:  0.20,
		model.ArmEarth: 0.20,
		model.ArmWind:  0.20,
		model.ArmWater: 0.20,
		model.ArmCore:  0.05,
		model.ArmVoid:  0.08,
		model.ArmOuter: 0.07,
	}

	var systemID int64 = 1
	for armID, ratio := range armDistribution {
		count := int(float64(systemCount) * ratio)
		for i := 0; i < count; i++ {
			sys := g.generateSystem(systemID, armID)
			result.Systems = append(result.Systems, sys)

			planets := g.generatePlanets(systemID, sys.StarType, sys.PlanetCount)
			result.Planets = append(result.Planets, planets...)

			belts := g.generateBelts(systemID, sys.BeltCount, armID)
			result.Belts = append(result.Belts, belts...)

			systemID++
		}
	}

	result.Gates = g.generateGateNetwork(result.Systems)

	return result
}

func (g *StarmapGenerator) generateSystem(id int64, armID model.ArmID) model.StarSystem {
	seed := g.rng.Int63()
	r := rand.New(rand.NewSource(seed))

	x, y, z := g.armCoordinates(armID, r)
	starType := g.rollStarType(armID, r)
	secLevel := g.rollSecurityLevel(armID, r)
	planetCount := r.Intn(8) + 1
	beltCount := r.Intn(4)
	hasAnomaly := r.Float64() < anomalyChance(armID)
	hasRuins := r.Float64() < ruinsChance(armID)

	armName := model.ArmNames[armID]
	name := fmt.Sprintf("%s-%04d", armName[:6], id)

	return model.StarSystem{
		ID:            id,
		Name:          name,
		ArmID:         armID,
		CoordX:        x,
		CoordY:        y,
		CoordZ:        z,
		SecurityLevel: secLevel,
		StarType:      starType,
		PlanetCount:   planetCount,
		BeltCount:     beltCount,
		HasAnomaly:    hasAnomaly,
		HasRuins:      hasRuins,
		Seed:          seed,
	}
}

func (g *StarmapGenerator) armCoordinates(armID model.ArmID, r *rand.Rand) (float64, float64, float64) {
	baseAngle := map[model.ArmID]float64{
		model.ArmFire:  0,
		model.ArmEarth: math.Pi / 2,
		model.ArmWind:  math.Pi,
		model.ArmWater: 3 * math.Pi / 2,
		model.ArmCore:  0,
		model.ArmVoid:  0,
		model.ArmOuter: 0,
	}

	switch armID {
	case model.ArmCore:
		dist := r.Float64() * 500
		angle := r.Float64() * 2 * math.Pi
		return dist * math.Cos(angle), dist * math.Sin(angle), (r.Float64() - 0.5) * 100

	case model.ArmVoid:
		angle := r.Float64() * 2 * math.Pi
		dist := 1000 + r.Float64()*3000
		offset := (r.Float64() - 0.5) * 800
		spiralAngle := angle + dist/1000
		return dist*math.Cos(spiralAngle) + offset, dist*math.Sin(spiralAngle) + offset, (r.Float64() - 0.5) * 200

	case model.ArmOuter:
		angle := r.Float64() * 2 * math.Pi
		dist := 4000 + r.Float64()*2000
		return dist * math.Cos(angle), dist * math.Sin(angle), (r.Float64() - 0.5) * 300

	default:
		angle := baseAngle[armID]
		dist := 500 + r.Float64()*3500
		spiralOffset := dist / 800
		armWidth := 200 + r.Float64()*400
		finalAngle := angle + spiralOffset
		x := dist*math.Cos(finalAngle) + armWidth*(r.Float64()-0.5)*math.Sin(finalAngle)
		y := dist*math.Sin(finalAngle) + armWidth*(r.Float64()-0.5)*math.Cos(finalAngle)
		z := (r.Float64() - 0.5) * 150
		return x, y, z
	}
}

func (g *StarmapGenerator) rollStarType(armID model.ArmID, r *rand.Rand) model.StarType {
	roll := r.Float64()

	if armID == model.ArmCore {
		switch {
		case roll < 0.15:
			return model.StarTypeBlackH
		case roll < 0.25:
			return model.StarTypeNeutron
		case roll < 0.35:
			return model.StarTypePulsar
		case roll < 0.55:
			return model.StarTypeO
		default:
			return model.StarTypeB
		}
	}

	if armID == model.ArmFire {
		switch {
		case roll < 0.10:
			return model.StarTypeO
		case roll < 0.25:
			return model.StarTypeB
		case roll < 0.40:
			return model.StarTypeA
		case roll < 0.60:
			return model.StarTypeF
		case roll < 0.80:
			return model.StarTypeG
		default:
			return model.StarTypeBinary
		}
	}

	switch {
	case roll < 0.02:
		return model.StarTypeO
	case roll < 0.08:
		return model.StarTypeB
	case roll < 0.18:
		return model.StarTypeA
	case roll < 0.30:
		return model.StarTypeF
	case roll < 0.50:
		return model.StarTypeG
	case roll < 0.70:
		return model.StarTypeK
	case roll < 0.90:
		return model.StarTypeM
	case roll < 0.95:
		return model.StarTypeBinary
	case roll < 0.98:
		return model.StarTypeNeutron
	default:
		return model.StarTypePulsar
	}
}

func (g *StarmapGenerator) rollSecurityLevel(armID model.ArmID, r *rand.Rand) float64 {
	switch armID {
	case model.ArmCore:
		return -0.1 - r.Float64()*0.9
	case model.ArmOuter:
		return -r.Float64() * 0.5
	case model.ArmVoid:
		return 0.0
	default:
		roll := r.Float64()
		switch {
		case roll < 0.10:
			return 0.5 + r.Float64()*0.5
		case roll < 0.25:
			return 0.1 + r.Float64()*0.3
		default:
			return 0.0
		}
	}
}

func anomalyChance(armID model.ArmID) float64 {
	switch armID {
	case model.ArmCore:
		return 0.60
	case model.ArmWater:
		return 0.30
	case model.ArmOuter:
		return 0.40
	default:
		return 0.15
	}
}

func ruinsChance(armID model.ArmID) float64 {
	switch armID {
	case model.ArmCore:
		return 0.50
	case model.ArmEarth:
		return 0.15
	default:
		return 0.08
	}
}

var planetTypes = []string{"熔岩行星", "岩石行星", "沙漠行星", "海洋行星", "生态行星", "冰封行星", "气态巨行星", "冰巨行星"}

func (g *StarmapGenerator) generatePlanets(systemID int64, starType model.StarType, count int) []model.Planet {
	r := rand.New(rand.NewSource(systemID * 31337))
	planets := make([]model.Planet, 0, count)

	for i := 0; i < count; i++ {
		orbitAU := 0.2 + float64(i)*0.7 + r.Float64()*0.5
		typeIdx := g.planetTypeByOrbit(orbitAU, r)

		planets = append(planets, model.Planet{
			SystemID:   systemID,
			Name:       fmt.Sprintf("P%d", i+1),
			PlanetType: planetTypes[typeIdx],
			OrbitAU:    math.Round(orbitAU*100) / 100,
			MoonCount:  r.Intn(4),
			HasStation: i == 0 && r.Float64() < 0.3,
		})
	}
	return planets
}

func (g *StarmapGenerator) planetTypeByOrbit(au float64, r *rand.Rand) int {
	switch {
	case au < 0.5:
		return 0 // 熔岩
	case au < 1.0:
		return 1 + r.Intn(2) // 岩石/沙漠
	case au < 2.0:
		return 2 + r.Intn(3) // 沙漠/海洋/生态
	case au < 4.0:
		return 5 + r.Intn(2) // 冰封/气态巨行星
	default:
		return 6 + r.Intn(2) // 气态巨行星/冰巨行星
	}
}

var beltTypes = []string{"普通矿带", "富矿带", "稀有矿带", "冰矿带", "异常矿带"}

func (g *StarmapGenerator) generateBelts(systemID int64, count int, armID model.ArmID) []model.AsteroidBelt {
	r := rand.New(rand.NewSource(systemID * 77773))
	belts := make([]model.AsteroidBelt, 0, count)

	for i := 0; i < count; i++ {
		typeIdx := 0
		roll := r.Float64()
		switch {
		case roll < 0.50:
			typeIdx = 0
		case roll < 0.75:
			typeIdx = 1
		case roll < 0.88:
			typeIdx = 3
		case roll < 0.96:
			typeIdx = 2
		default:
			typeIdx = 4
		}
		if armID == model.ArmEarth && typeIdx < 2 {
			typeIdx++ // 厚土之臂矿带品质+1
		}

		belts = append(belts, model.AsteroidBelt{
			SystemID:  systemID,
			Name:      fmt.Sprintf("矿带-%c", 'A'+i),
			BeltType:  beltTypes[typeIdx],
			OrbitAU:   1.5 + float64(i)*1.2 + r.Float64(),
			Remaining: 100,
		})
	}
	return belts
}

func (g *StarmapGenerator) generateGateNetwork(systems []model.StarSystem) []model.Stargate {
	var gates []model.Stargate
	var gateID int64 = 1

	for i := range systems {
		neighbors := g.findNearestSystems(systems, i, 3)
		for _, j := range neighbors {
			if !gateExists(gates, systems[i].ID, systems[j].ID) {
				gates = append(gates, model.Stargate{
					ID: gateID, FromID: systems[i].ID, ToID: systems[j].ID, IsNatural: true,
				})
				gateID++
				gates = append(gates, model.Stargate{
					ID: gateID, FromID: systems[j].ID, ToID: systems[i].ID, IsNatural: true,
				})
				gateID++
			}
		}
	}
	return gates
}

func (g *StarmapGenerator) findNearestSystems(systems []model.StarSystem, idx int, count int) []int {
	type distIdx struct {
		dist float64
		idx  int
	}

	src := systems[idx]
	var candidates []distIdx

	for i := range systems {
		if i == idx {
			continue
		}
		dx := src.CoordX - systems[i].CoordX
		dy := src.CoordY - systems[i].CoordY
		dz := src.CoordZ - systems[i].CoordZ
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		candidates = append(candidates, distIdx{dist, i})
	}

	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].dist < candidates[i].dist {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	result := make([]int, 0, count)
	for i := 0; i < count && i < len(candidates); i++ {
		result = append(result, candidates[i].idx)
	}
	return result
}

func gateExists(gates []model.Stargate, from, to int64) bool {
	for _, g := range gates {
		if (g.FromID == from && g.ToID == to) || (g.FromID == to && g.ToID == from) {
			return true
		}
	}
	return false
}
