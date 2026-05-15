package ui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"deckbuilder/combat"
	"deckbuilder/enemies"
	"deckbuilder/runes"
)

const (
	ScreenW = 1280
	ScreenH = 800

	RadarCX, RadarCY = 480, 360
	RadarRadius      = 320

	HandY      = 640
	CardW      = 110
	CardH      = 150
	CardGap    = 12
	CardStartX = 80

	EndTurnX, EndTurnY  = 1100, 660
	EndTurnW, EndTurnH = 140, 60

	DashBtnX, DashBtnY  = 1100, 590
	DashBtnW, DashBtnH = 140, 60

	StageY      = 470
	StageCardW  = 100
	StageCardH  = 100
	StageGap    = 10
	StageStartX = 80

	// Reward screen layout
	RewardCardW    = 220
	RewardCardH    = 300
	RewardGap      = 40
	RewardTopY     = 220
	SkipBtnW, SkipBtnH = 220, 50
)

// RunView is the rendering surface — what game/Run exposes to the UI without
// the UI depending on the game package.
type RunView struct {
	Phase           int
	Class           runes.Class
	Combat          *combat.Combat
	Rewards         []runes.Card
	EncounterIdx    int
	TotalEncounters int
	PlayerHP, MaxHP int
	DeckSize        int
}

// Run phase constants kept in sync with game.RunPhase.
const (
	phaseSelectClass = 0
	phaseInCombat    = 1
	phaseReward      = 2
	phaseRunWon      = 3
	phaseRunLost     = 4
)

var (
	bgColor      = color.RGBA{18, 18, 28, 255}
	radarColor   = color.RGBA{40, 50, 80, 255}
	radarRing    = color.RGBA{70, 90, 130, 255}
	playerColor  = color.RGBA{120, 220, 255, 255}
	enemyColor   = color.RGBA{230, 90, 90, 255}
	enemyDeadCol = color.RGBA{60, 60, 60, 255}
	cardBg       = color.RGBA{50, 40, 70, 255}
	cardBgHi     = color.RGBA{90, 70, 130, 255}
	cardBgDim    = color.RGBA{30, 25, 40, 255}
	white        = color.RGBA{240, 240, 240, 255}
	yellow       = color.RGBA{240, 220, 100, 255}
	green        = color.RGBA{120, 220, 120, 255}
	moveOkCol    = color.RGBA{120, 220, 140, 220}
	moveOverCol  = color.RGBA{230, 160, 90, 220}
	moveRingCol  = color.RGBA{120, 220, 140, 90}
	tooltipBg    = color.RGBA{20, 20, 32, 240}
	tooltipEdge  = color.RGBA{120, 130, 170, 255}

	fireCol     = color.RGBA{255, 150, 70, 255}
	frostCol    = color.RGBA{120, 200, 255, 255}
	physicalCol = color.RGBA{230, 230, 230, 255}
	minionCol   = color.RGBA{120, 220, 160, 255}
	cardBorderCol  = color.RGBA{18, 16, 26, 255}
	titleBarCol    = color.RGBA{32, 28, 44, 255}
	titleSepCol    = color.RGBA{80, 70, 100, 255}
	glyphGoldCol   = color.RGBA{200, 170, 90, 255}
	glyphSilverCol = color.RGBA{170, 175, 190, 255}
	glyphTextCol   = color.RGBA{30, 25, 15, 255}
	costGemCol     = color.RGBA{90, 110, 200, 255}
	descFgCol      = color.RGBA{210, 200, 230, 255}

	wallCol         = color.RGBA{170, 160, 140, 255}
	wallOuterCol    = color.RGBA{60, 55, 48, 255}
	wallHighlightCol = color.RGBA{210, 200, 180, 230}
	wallCrackCol    = color.RGBA{30, 25, 22, 230}
	slowCol        = color.RGBA{240, 200, 90, 255}
	rangeRingCol   = color.RGBA{180, 130, 220, 110}
	rangeTargetCol = color.RGBA{255, 200, 130, 220}
	coneCol            = color.RGBA{120, 220, 255, 18}
	coneEdgeCol        = color.RGBA{120, 220, 255, 60}
	predictCol         = color.RGBA{230, 90, 90, 110}
	enemyOutOfConeCol  = color.RGBA{140, 70, 70, 210}
	hitFlashEnemyCol   = color.RGBA{255, 240, 220, 255}
	hitFlashMinionCol  = color.RGBA{230, 255, 230, 255}
	hitFlashPlayerCol  = color.RGBA{255, 200, 200, 255}

	logBg          = color.RGBA{16, 18, 26, 220}
	logEdge        = color.RGBA{70, 80, 110, 255}
	logColPlayer   = color.RGBA{180, 230, 200, 255}
	logColEnemy    = color.RGBA{240, 160, 150, 255}
	logColMinion   = color.RGBA{170, 220, 220, 255}
	logColSystem   = color.RGBA{170, 170, 180, 255}
)

const (
	LogPanelX = 880
	LogPanelY = 160
	LogPanelW = 380
	LogPanelH = 480
)

func logKindColor(k combat.LogKind) color.RGBA {
	switch k {
	case combat.LogPlayer:
		return logColPlayer
	case combat.LogEnemy:
		return logColEnemy
	case combat.LogMinion:
		return logColMinion
	default:
		return logColSystem
	}
}

// cardKind selects layout-density preset for drawCardFrame.
type cardKind int

const (
	cardKindHand cardKind = iota
	cardKindStage
	cardKindReward
)

// drawCardFrame paints a structured rune card: outer dark border, title bar
// with a glyph disc (gold for Core, silver for Modifier) and cost gem, then a
// body area sized to the kind. bg is the inner panel fill; border is the
// outer ring color (lets the caller mark Slow runes with a colored border).
func drawCardFrame(screen *ebiten.Image, card runes.Card, x, y, w, h int, bg, border color.RGBA, kind cardKind) {
	var titleH, glyphR, lineH, maxLines, pad int
	var nameFace, descFace, glyphFace fontFace
	switch kind {
	case cardKindStage:
		titleH, glyphR, pad = 22, 7, 6
		nameFace, descFace, glyphFace = faceSmall, faceSmall, faceBody
		maxLines = 0
	case cardKindReward:
		titleH, glyphR, pad = 40, 16, 10
		nameFace, descFace, glyphFace = faceBody, faceSmall, faceTitle
		lineH = 16
		maxLines = 20
	default: // cardKindHand
		titleH, glyphR, pad = 22, 8, 6
		nameFace, descFace, glyphFace = faceSmall, faceSmall, faceBody
		lineH = 14
		maxLines = 5
	}

	// Outer border.
	vector.DrawFilledRect(screen, float32(x-1), float32(y-1), float32(w+2), float32(h+2), border, true)
	// Body fill.
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bg, true)
	// Title bar.
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(titleH), titleBarCol, true)
	vector.StrokeLine(screen, float32(x), float32(y+titleH), float32(x+w), float32(y+titleH), 1, titleSepCol, true)

	// Glyph disc, gold for Core, silver for Modifier.
	glyphFill := glyphGoldCol
	if card.Role == runes.RoleModifier {
		glyphFill = glyphSilverCol
	}
	gcx := x + pad + glyphR
	gcy := y + titleH/2
	vector.DrawFilledCircle(screen, float32(gcx), float32(gcy), float32(glyphR), glyphFill, true)
	// Outline ring on the glyph disc for separation.
	vector.StrokeCircle(screen, float32(gcx), float32(gcy), float32(glyphR), 1, cardBorderCol, true)
	centerText(screen, card.Glyph, gcx-glyphR, gcy-glyphFace.ascent/2, glyphR*2, glyphFace, glyphTextCol)

	// Cost gem.
	costR := glyphR
	ccx := x + w - pad - costR
	ccy := y + titleH/2
	vector.DrawFilledCircle(screen, float32(ccx), float32(ccy), float32(costR), costGemCol, true)
	vector.StrokeCircle(screen, float32(ccx), float32(ccy), float32(costR), 1, cardBorderCol, true)
	centerText(screen, fmt.Sprintf("%d", card.Cost), ccx-costR, ccy-nameFace.ascent/2, costR*2, nameFace, white)

	// Name centered between glyph and cost.
	nameX := gcx + glyphR + 4
	nameW := ccx - costR - 4 - nameX
	if nameW > 0 {
		centerText(screen, truncateToWidth(card.Name, nameFace, nameW), nameX, y+titleH/2-nameFace.ascent/2, nameW, nameFace, white)
	}

	// Body description (skipped for the compact stage layout).
	if maxLines > 0 {
		descX := x + pad
		descY := y + titleH + pad
		descW := w - 2*pad
		lines := wrapText(card.Description, float64(descW), descFace)
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines[maxLines-1] = truncateToWidth(lines[maxLines-1]+"…", descFace, descW)
		}
		for i, line := range lines {
			drawText(screen, line, descX, descY+i*lineH, descFace, descFgCol)
		}
	}

	// Footer tags.
	footY := y + h - 18
	if card.Slow {
		drawText(screen, "slow", x+pad, footY, faceSmall, slowCol)
	}
	if card.Range > 0 {
		rstr := fmt.Sprintf("range %.0f", card.Range)
		rW := int(textWidth(rstr, faceSmall))
		drawText(screen, rstr, x+w-pad-rW, footY, faceSmall, rangeRingCol)
	}
}

// truncateToWidth returns s shortened with an ellipsis until it fits maxW.
func truncateToWidth(s string, ff fontFace, maxW int) string {
	if textWidth(s, ff) <= float64(maxW) {
		return s
	}
	const ell = "…"
	for len(s) > 0 {
		s = s[:len(s)-1]
		if textWidth(s+ell, ff) <= float64(maxW) {
			return s + ell
		}
	}
	return ell
}

func damageTypeColor(t runes.DamageType) color.RGBA {
	switch t {
	case runes.Fire:
		return fireCol
	case runes.Frost:
		return frostCol
	default:
		return physicalCol
	}
}

func DrawRun(screen *ebiten.Image, v RunView) {
	screen.Fill(bgColor)
	switch v.Phase {
	case phaseSelectClass:
		drawClassSelect(screen)
	case phaseInCombat:
		drawCombatScreen(screen, v)
	case phaseReward:
		drawCombatScreen(screen, v) // backdrop
		drawRewardOverlay(screen, v)
	case phaseRunWon:
		drawCombatScreen(screen, v)
		drawEndOverlay(screen, v, true)
	case phaseRunLost:
		drawCombatScreen(screen, v)
		drawEndOverlay(screen, v, false)
	}
}

func drawCombatScreen(screen *ebiten.Image, v RunView) {
	c := v.Combat
	drawRadar(screen, c)
	drawBeams(screen, c)
	drawCastPulse(screen, c)
	drawParticles(screen, c)
	drawRangePreview(screen, c)
	drawPopups(screen, c)
	drawPlacementPreview(screen, c)
	drawStage(screen, c)
	drawHand(screen, c)
	drawHUD(screen, v)
	drawLog(screen, c)
	drawDashBtn(screen, c)
	drawEndTurn(screen, c)
	drawCardTooltip(screen, c)
	drawPhaseBanner(screen, c)
}

// drawLog renders the most recent combat events along the right side. Newest
// at the bottom — chronological reading order.
func drawLog(screen *ebiten.Image, c *combat.Combat) {
	vector.DrawFilledRect(screen, LogPanelX, LogPanelY, LogPanelW, LogPanelH, logBg, true)
	vector.StrokeRect(screen, LogPanelX, LogPanelY, LogPanelW, LogPanelH, 1, logEdge, true)
	drawText(screen, "Combat log", LogPanelX+8, LogPanelY+4, faceSmall, white)

	const (
		lineH = 16
		topY  = LogPanelY + 24
		botY  = LogPanelY + LogPanelH - 8
	)
	visible := (botY - topY) / lineH
	start := len(c.Log) - visible
	if start < 0 {
		start = 0
	}
	for i, e := range c.Log[start:] {
		y := topY + i*lineH
		drawColoredText(screen, e.Text, LogPanelX+8, y, logKindColor(e.Kind), 1)
	}
}

// drawRangePreview draws the range circle and likely target for the card the
// cursor is hovering in the player's hand. Cards with no Range or that aren't
// playable show nothing.
func drawRangePreview(screen *ebiten.Image, c *combat.Combat) {
	if c.PendingCardIdx >= 0 {
		return
	}
	mx, my := ebiten.CursorPosition()
	i := HitCard(c, mx, my)
	if i < 0 {
		return
	}
	card := c.Hand[i]
	if card.Range <= 0 {
		return
	}
	vector.StrokeCircle(screen, RadarCX, RadarCY, float32(card.Range), 1, rangeRingCol, true)

	// Highlight the enemy that would be hit (nearest in range with LOS).
	if target := nearestInRangeForPreview(c, card.Range); target != nil {
		ex := float32(RadarCX + (target.X - c.Player.X))
		ey := float32(RadarCY + (target.Y - c.Player.Y))
		vector.StrokeCircle(screen, ex, ey, 14, 2, rangeTargetCol, true)
	}
}

// nearestInRangeForPreview is a UI-only helper that mirrors the targeting
// rules in combat (nearest enemy within range with line-of-sight from player).
// We intentionally walk c.Enemies directly rather than expose more API.
func nearestInRangeForPreview(c *combat.Combat, r float64) *combatEnemyView {
	var bestX, bestY float64
	bestD := math.Inf(1)
	found := false
	for _, e := range c.Enemies {
		if e.HP <= 0 {
			continue
		}
		d := math.Hypot(e.X-c.Player.X, e.Y-c.Player.Y)
		if d > r {
			continue
		}
		if !inConePreview(c, e.X, e.Y) {
			continue
		}
		blocked := false
		for _, w := range c.Walls {
			if w.HP <= 0 {
				continue
			}
			if hit, _, _, _ := segIntersectFor(c.Player.X, c.Player.Y, e.X, e.Y, w.X1, w.Y1, w.X2, w.Y2); hit {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		if d < bestD {
			bestD = d
			bestX = e.X
			bestY = e.Y
			found = true
		}
	}
	if !found {
		return nil
	}
	return &combatEnemyView{X: bestX, Y: bestY}
}

// inConePreview mirrors combat.Combat.inCone for the UI's hover preview.
func inConePreview(c *combat.Combat, tx, ty float64) bool {
	dx := tx - c.Player.X
	dy := ty - c.Player.Y
	if dx == 0 && dy == 0 {
		return true
	}
	angle := math.Atan2(dy, dx)
	delta := angle - c.Player.Facing
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta < -math.Pi {
		delta += 2 * math.Pi
	}
	if delta < 0 {
		delta = -delta
	}
	return delta <= combat.ConeHalfAngle
}

type combatEnemyView struct {
	X, Y float64
}

// segIntersectFor is a tiny duplicate of combat's segment intersection used
// for UI LOS checks. Avoids exposing more package internals just for hover.
func segIntersectFor(ax, ay, bx, by, cx, cy, dx, dy float64) (bool, float64, float64, float64) {
	rX := bx - ax
	rY := by - ay
	sX := dx - cx
	sY := dy - cy
	denom := rX*sY - rY*sX
	if denom == 0 {
		return false, 0, 0, 0
	}
	t := ((cx-ax)*sY - (cy-ay)*sX) / denom
	u := ((cx-ax)*rY - (cy-ay)*rX) / denom
	if t < 0 || t > 1 || u < 0 || u > 1 {
		return false, 0, 0, 0
	}
	return true, t, ax + t*rX, ay + t*rY
}

func drawStage(screen *ebiten.Image, c *combat.Combat) {
	header := "Spell stage (right-click to unstage):"
	if c.SpellCast {
		header = "Spell cast — press E to end turn"
	} else if len(c.Stage) == 0 {
		header = "Spell stage:  (click a rune to add it)"
	}
	drawText(screen, header, StageStartX, StageY-20, faceSmall, white)

	for i, sc := range c.Stage {
		x := StageStartX + i*(StageCardW+StageGap)
		y := StageY
		vector.DrawFilledRect(screen, float32(x), float32(y), StageCardW, StageCardH, cardBg, true)
		border := cardBorderCol
		if sc.Card.Slow {
			border = slowCol
		}
		drawCardFrame(screen, sc.Card, x, y, StageCardW, StageCardH, cardBg, border, cardKindStage)
	}
	if !c.SpellCast && len(c.Stage) > 0 {
		total := 0
		for _, sc := range c.Stage {
			total += sc.Card.Cost
		}
		castVerb := "press E to cast"
		if c.StageIsSlow() {
			castVerb = "press E to cast — SLOW (resolves after enemies)"
		}
		summary := fmt.Sprintf("%d rune(s), %d energy — %s", len(c.Stage), total, castVerb)
		drawText(screen, summary, StageStartX, StageY+StageCardH+8, faceSmall, white)
	}
}

func drawPlacementPreview(screen *ebiten.Image, c *combat.Combat) {
	if c.PendingCardIdx < 0 || c.PendingCardIdx >= len(c.Hand) {
		return
	}
	card := c.Hand[c.PendingCardIdx]
	mx, my := ebiten.CursorPosition()
	if rx, ry, ok := HitRadar(mx, my); ok {
		gx := float32(RadarCX + rx)
		gy := float32(RadarCY + ry)
		switch card.PlacementShape {
		case runes.PlacementWall:
			// perpendicular preview through cursor (length 100, matches card)
			d := math.Hypot(rx, ry)
			if d == 0 {
				rx, ry, d = 1, 0, 1
			}
			pxv := float32(-ry / d)
			pyv := float32(rx / d)
			const half = float32(50)
			vector.StrokeLine(screen, gx-pxv*half, gy-pyv*half, gx+pxv*half, gy+pyv*half, 4, wallCol, true)
			vector.StrokeLine(screen, RadarCX, RadarCY, gx, gy, 1, minionCol, true)
		default:
			vector.StrokeCircle(screen, gx, gy, 8, 2, minionCol, true)
			vector.StrokeLine(screen, RadarCX, RadarCY, gx, gy, 1, minionCol, true)
		}
	}
	banner := fmt.Sprintf("Placing: %s — left-click radar to confirm, right-click to cancel", card.Name)
	drawText(screen, banner, 40, 30, faceSmall, white)
}

func drawColoredText(screen *ebiten.Image, s string, x, y int, c color.RGBA, alpha float32) {
	clr := color.RGBA{
		R: uint8(float32(c.R) * alpha),
		G: uint8(float32(c.G) * alpha),
		B: uint8(float32(c.B) * alpha),
		A: uint8(255 * alpha),
	}
	drawText(screen, s, x, y, faceSmall, clr)
}

// drawBeams renders short-lived spell impact beams in world space, faded by
// remaining life. Drawn over the dungeon floor but under entities.
func drawBeams(screen *ebiten.Image, c *combat.Combat) {
	if len(c.Beams) == 0 {
		return
	}
	playerDX, playerDY := c.PlayerDisplayPos()
	sx, sy := c.ViewShake()
	for _, b := range c.Beams {
		t := b.Life / b.MaxLife
		if t > 1 {
			t = 1
		} else if t < 0 {
			continue
		}
		a := uint8(200 * t)
		col := color.RGBA{b.R, b.G, b.B, a}
		x1 := float32(float64(RadarCX) + (b.X1 - playerDX) + sx)
		y1 := float32(float64(RadarCY) + (b.Y1 - playerDY) + sy)
		x2 := float32(float64(RadarCX) + (b.X2 - playerDX) + sx)
		y2 := float32(float64(RadarCY) + (b.Y2 - playerDY) + sy)
		// Outer thicker faded line + brighter thin core.
		vector.StrokeLine(screen, x1, y1, x2, y2, 4, color.RGBA{b.R, b.G, b.B, a / 3}, true)
		vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, col, true)
	}
}

// drawCastPulse renders the expanding ring around the player when a spell is
// cast or a dash is performed. Decays linearly over CastPulseDuration.
func drawCastPulse(screen *ebiten.Image, c *combat.Combat) {
	if c.CastPulseTimer <= 0 {
		return
	}
	p := 1 - c.CastPulseTimer/combat.CastPulseDuration // 0 → fresh, 1 → done
	if p < 0 {
		p = 0
	} else if p > 1 {
		p = 1
	}
	radius := float32(10 + p*60)
	a := uint8(220 * (1 - p))
	col := color.RGBA{c.CastPulseR, c.CastPulseG, c.CastPulseB, a}
	sx, sy := c.ViewShake()
	pcx := float32(RadarCX) + float32(sx)
	pcy := float32(RadarCY) + float32(sy)
	vector.StrokeCircle(screen, pcx, pcy, radius, 2, col, true)
}

// drawParticles renders the active particle bursts at world-space positions,
// translated through the lerped player offset and current view shake.
func drawParticles(screen *ebiten.Image, c *combat.Combat) {
	if len(c.Particles) == 0 {
		return
	}
	playerDX, playerDY := c.PlayerDisplayPos()
	sx, sy := c.ViewShake()
	for _, p := range c.Particles {
		t := p.Life / p.MaxLife
		if t > 1 {
			t = 1
		} else if t < 0 {
			t = 0
		}
		a := uint8(255 * t)
		col := color.RGBA{p.R, p.G, p.B, a}
		x := float32(float64(RadarCX) + (p.X - playerDX) + sx)
		y := float32(float64(RadarCY) + (p.Y - playerDY) + sy)
		vector.DrawFilledCircle(screen, x, y, p.Size, col, true)
	}
}

func drawPopups(screen *ebiten.Image, c *combat.Combat) {
	playerDX, playerDY := c.PlayerDisplayPos()
	sx, sy := c.ViewShake()
	for _, p := range c.Popups {
		t := p.Age / combat.PopupLife
		if t > 1 {
			continue
		}
		alpha := float32(1.0 - t*t) // ease-out fade
		dy := -float64(40) * t
		x := int(float64(RadarCX) + (p.X - playerDX) + sx)
		y := int(float64(RadarCY) + (p.Y - playerDY) + dy - 10 + sy)
		drawColoredText(screen, fmt.Sprintf("%d", p.Amount), x-6, y, damageTypeColor(p.Type), alpha)
	}
}

func drawRadar(screen *ebiten.Image, c *combat.Combat) {
	sx, sy := c.ViewShake()
	shakeX := float32(sx)
	shakeY := float32(sy)
	playerDX, playerDY := c.PlayerDisplayPos()

	drawDungeonFloor(screen)
	for _, r := range []float32{RadarRadius * 0.33, RadarRadius * 0.66, RadarRadius} {
		vector.StrokeCircle(screen, RadarCX, RadarCY, r, 1, radarRing, true)
	}
	drawCone(screen, c)
	drawMovePreview(screen, c)

	// Player dot + facing tick (flashes when hit; shakes with the world).
	playerCol := playerColor
	if c.PlayerHitTimer > 0 {
		playerCol = hitFlashPlayerCol
	}
	pcx := float32(RadarCX) + shakeX
	pcy := float32(RadarCY) + shakeY
	vector.DrawFilledCircle(screen, pcx, pcy, 8, playerCol, true)
	tx := pcx + float32(math.Cos(c.Player.Facing)*16)
	ty := pcy + float32(math.Sin(c.Player.Facing)*16)
	vector.StrokeLine(screen, pcx, pcy, tx, ty, 3, playerCol, true)

	for _, w := range c.Walls {
		if w.HP <= 0 {
			continue
		}
		x1 := float32(RadarCX+(w.X1-playerDX)) + shakeX
		y1 := float32(RadarCY+(w.Y1-playerDY)) + shakeY
		x2 := float32(RadarCX+(w.X2-playerDX)) + shakeX
		y2 := float32(RadarCY+(w.Y2-playerDY)) + shakeY
		// Layered stone: dark outer band, lighter stone body, thin highlight.
		vector.StrokeLine(screen, x1, y1, x2, y2, 8, wallOuterCol, true)
		vector.StrokeLine(screen, x1, y1, x2, y2, 5, wallCol, true)
		vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, wallHighlightCol, true)
		// Cracks proportional to damage taken.
		hpFrac := float64(w.HP) / float64(w.MaxHP)
		cracks := int((1 - hpFrac) * 6)
		if cracks > 0 {
			dx := float64(x2 - x1)
			dy := float64(y2 - y1)
			L := math.Hypot(dx, dy)
			if L > 0 {
				px := -dy / L * 4
				py := dx / L * 4
				for i := 0; i < cracks; i++ {
					t := (float64(i) + 0.5) / 6
					cx := float64(x1) + dx*t
					cy := float64(y1) + dy*t
					vector.StrokeLine(screen,
						float32(cx-px), float32(cy-py),
						float32(cx+px), float32(cy+py),
						1, wallCrackCol, true)
				}
			}
		}
		mx := int((x1 + x2) / 2)
		my := int((y1 + y2) / 2)
		drawText(screen, fmt.Sprintf("wall %d/%d", w.HP, w.MaxHP), mx-24, my-18, faceSmall, wallCol)
	}

	for _, e := range c.Enemies {
		ewx, ewy := c.EnemyDisplayPos(e)
		ex := float32(RadarCX+(ewx-playerDX)) + shakeX
		ey := float32(RadarCY+(ewy-playerDY)) + shakeY

		if e.HP > 0 {
			if px, py, moves := c.PredictMove(e); moves {
				gx := float32(RadarCX+(px-playerDX)) + shakeX
				gy := float32(RadarCY+(py-playerDY)) + shakeY
				vector.StrokeLine(screen, ex, ey, gx, gy, 1, predictCol, true)
				vector.StrokeCircle(screen, gx, gy, 9, 1, predictCol, true)
			}
		}

		col := enemyColor
		switch {
		case e.HP <= 0:
			col = enemyDeadCol
		case e.HitTimer > 0:
			col = hitFlashEnemyCol
		case !inConePreview(c, e.X, e.Y):
			col = enemyOutOfConeCol
		}
		drawEnemySilhouette(screen, e, ex, ey, col)
		label := fmt.Sprintf("%s %d/%d", e.Name, e.HP, e.MaxHP)
		drawText(screen, label, int(ex)-30, int(ey)+14, faceSmall, col)
		if e.HP > 0 && e.Intent != "" {
			drawText(screen, "intent: "+e.Intent, int(ex)-30, int(ey)+28, faceSmall, white)
		}
	}

	for _, m := range c.Minions {
		if m.HP <= 0 {
			continue
		}
		mx := float32(RadarCX+(m.X-playerDX)) + shakeX
		my := float32(RadarCY+(m.Y-playerDY)) + shakeY
		minionDrawCol := minionCol
		if m.HitTimer > 0 {
			minionDrawCol = hitFlashMinionCol
		}
		vector.DrawFilledCircle(screen, mx, my, 7, minionDrawCol, true)
		label := fmt.Sprintf("M %d/%d  (%d/t)", m.HP, m.MaxHP, m.AttackPower)
		drawText(screen, label, int(mx)-32, int(my)+10, faceSmall, minionCol)
	}
}

// --- Dungeon background ---

var dungeonBg *ebiten.Image

// ensureDungeonBg builds a pre-rendered radar-floor texture on first call:
// the base radar disc, sprinkled with darker stone speckles and a few thin
// rune-crack scratches. Deterministic via a fixed RNG seed so it doesn't
// shimmer between launches.
func ensureDungeonBg() {
	if dungeonBg != nil {
		return
	}
	size := RadarRadius * 2
	dungeonBg = ebiten.NewImage(size, size)
	center := float32(RadarRadius)
	vector.DrawFilledCircle(dungeonBg, center, center, float32(RadarRadius), radarColor, true)

	rng := rand.New(rand.NewSource(42))
	stoneCol := color.RGBA{25, 28, 42, 255}
	for i := 0; i < 110; i++ {
		// uniform sampling inside the disc: sqrt(rand) for radius
		r := math.Sqrt(rng.Float64()) * float64(RadarRadius-4)
		a := rng.Float64() * math.Pi * 2
		px := float32(float64(RadarRadius) + r*math.Cos(a))
		py := float32(float64(RadarRadius) + r*math.Sin(a))
		s := float32(rng.Float64()*1.5 + 0.5)
		vector.DrawFilledCircle(dungeonBg, px, py, s, stoneCol, true)
	}
	crackCol := color.RGBA{32, 36, 52, 255}
	for i := 0; i < 14; i++ {
		r := math.Sqrt(rng.Float64()) * float64(RadarRadius-20)
		a := rng.Float64() * math.Pi * 2
		cx := float64(RadarRadius) + r*math.Cos(a)
		cy := float64(RadarRadius) + r*math.Sin(a)
		length := rng.Float64()*14 + 6
		ca := rng.Float64() * math.Pi * 2
		ex := cx + math.Cos(ca)*length
		ey := cy + math.Sin(ca)*length
		vector.StrokeLine(dungeonBg, float32(cx), float32(cy), float32(ex), float32(ey), 1, crackCol, true)
	}
}

// drawDungeonFloor draws the pre-rendered dungeon texture at the radar
// position. Replaces the flat-filled circle that used to sit there.
func drawDungeonFloor(screen *ebiten.Image) {
	ensureDungeonBg()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(RadarCX-RadarRadius), float64(RadarCY-RadarRadius))
	screen.DrawImage(dungeonBg, op)
}

// drawEnemySilhouette renders a small distinctive shape per enemy type.
// Without external sprites, this just composes 2-4 vector primitives —
// enough to read "this is a Goblin vs a Wraith vs a Troll" at a glance.
// When the enemy is dead, we collapse to a single grey blob.
func drawEnemySilhouette(screen *ebiten.Image, e *enemies.Enemy, ex, ey float32, col color.RGBA) {
	if e.HP <= 0 {
		vector.DrawFilledCircle(screen, ex, ey, 9, col, true)
		return
	}
	eyeCol := color.RGBA{40, 10, 10, 255}
	switch e.Name {
	case "Goblin":
		vector.DrawFilledCircle(screen, ex, ey, 9, col, true)
		vector.DrawFilledCircle(screen, ex-3, ey-2, 1.4, eyeCol, true)
		vector.DrawFilledCircle(screen, ex+3, ey-2, 1.4, eyeCol, true)
	case "Wraith":
		// Outer haze, semi-transparent body, bright inner core.
		haze := color.RGBA{col.R, col.G, col.B, 70}
		bodyHi := color.RGBA{col.R, col.G, col.B, 200}
		coreCol := color.RGBA{255, 220, 230, 200}
		vector.DrawFilledCircle(screen, ex, ey, 14, haze, true)
		vector.DrawFilledCircle(screen, ex, ey, 11, bodyHi, true)
		vector.DrawFilledCircle(screen, ex, ey, 5, coreCol, true)
	case "Troll":
		// Bigger body with two horn dots and eyes.
		hornCol := color.RGBA{170, 140, 90, 255}
		vector.DrawFilledCircle(screen, ex, ey, 14, col, true)
		vector.DrawFilledCircle(screen, ex-7, ey-11, 3, hornCol, true)
		vector.DrawFilledCircle(screen, ex+7, ey-11, 3, hornCol, true)
		vector.DrawFilledCircle(screen, ex-4, ey-2, 1.6, eyeCol, true)
		vector.DrawFilledCircle(screen, ex+4, ey-2, 1.6, eyeCol, true)
	default:
		vector.DrawFilledCircle(screen, ex, ey, 10, col, true)
	}
}

// drawCone renders only the two boundary rays of the player's targeting cone.
// No fill — keeps the radar readable. Facing tick on the player dot conveys
// the direction; these rays just mark where the cone ends.
func drawCone(screen *ebiten.Image, c *combat.Combat) {
	const radius = float64(RadarRadius)
	half := combat.ConeHalfAngle
	facing := c.Player.Facing
	ax := float32(RadarCX) + float32(math.Cos(facing-half)*radius)
	ay := float32(RadarCY) + float32(math.Sin(facing-half)*radius)
	bx := float32(RadarCX) + float32(math.Cos(facing+half)*radius)
	by := float32(RadarCY) + float32(math.Sin(facing+half)*radius)
	vector.StrokeLine(screen, RadarCX, RadarCY, ax, ay, 1, coneEdgeCol, true)
	vector.StrokeLine(screen, RadarCX, RadarCY, bx, by, 1, coneEdgeCol, true)
}

func drawMovePreview(screen *ebiten.Image, c *combat.Combat) {
	if c.Phase != combat.PhasePlayer {
		return
	}
	if c.MovementBudget <= 0 {
		return
	}
	// Budget ring around the player.
	vector.StrokeCircle(screen, RadarCX, RadarCY, float32(c.MovementBudget), 1, moveRingCol, true)

	mx, my := ebiten.CursorPosition()
	rx, ry, ok := HitRadar(mx, my)
	if !ok {
		return
	}
	dist := math.Hypot(rx, ry)
	if dist < 1 {
		return
	}
	inBudget := dist <= c.MovementBudget
	step := dist
	if !inBudget {
		step = c.MovementBudget
	}
	gx := float32(RadarCX + rx/dist*step)
	gy := float32(RadarCY + ry/dist*step)

	col := moveOkCol
	if !inBudget {
		col = moveOverCol
	}
	vector.StrokeLine(screen, RadarCX, RadarCY, gx, gy, 2, col, true)
	vector.StrokeCircle(screen, gx, gy, 8, 2, col, true)

	label := fmt.Sprintf("%.0f / %.0f", dist, c.MovementBudget)
	if !inBudget {
		label += "  (clamped)"
	}
	drawText(screen, label, int(gx)+12, int(gy)-6, faceSmall, col)
}

func drawHand(screen *ebiten.Image, c *combat.Combat) {
	mx, my := ebiten.CursorPosition()
	for i, card := range c.Hand {
		x, y := cardRect(i)
		playable, _ := cardPlayable(c, card)
		bg := cardBg
		border := cardBorderCol
		if i == c.PendingCardIdx {
			bg = cardBgHi
			border = slowCol
		} else if !playable {
			bg = cardBgDim
		} else if mx >= x && mx < x+CardW && my >= y && my < y+CardH {
			bg = cardBgHi
		}
		if card.Slow {
			border = slowCol
		}
		drawCardFrame(screen, card, x, y, CardW, CardH, bg, border, cardKindHand)
	}
}

func cardRect(i int) (int, int) {
	return CardStartX + i*(CardW+CardGap), HandY
}

// cardPlayable evaluates whether a hand card can be staged right now,
// accounting for spell-already-cast, energy, composition, and CanPlay
// constraints. Energy is checked against the post-staging projection
// (Energy - already-staged cost).
func cardPlayable(c *combat.Combat, card runes.Card) (bool, string) {
	if c.SpellCast {
		return false, "spell already cast this turn"
	}
	if c.TotalStagedCost()+card.Cost > c.Energy {
		return false, "not enough energy"
	}
	if ok, why := c.CanAddToStage(card); !ok {
		return false, why
	}
	if card.CanPlay != nil {
		if ok, why := card.CanPlay(c); !ok {
			return false, why
		}
	}
	return true, ""
}

// HitCard returns the index of the card at (mx,my), or -1.
func HitCard(c *combat.Combat, mx, my int) int {
	for i := range c.Hand {
		x, y := cardRect(i)
		if mx >= x && mx < x+CardW && my >= y && my < y+CardH {
			return i
		}
	}
	return -1
}

// HitRadar returns radar-relative coordinates if (mx,my) is inside the radar.
func HitRadar(mx, my int) (rx, ry float64, ok bool) {
	dx := float64(mx - RadarCX)
	dy := float64(my - RadarCY)
	if math.Hypot(dx, dy) > RadarRadius {
		return 0, 0, false
	}
	return dx, dy, true
}

func HitEndTurn(mx, my int) bool {
	return mx >= EndTurnX && mx < EndTurnX+EndTurnW &&
		my >= EndTurnY && my < EndTurnY+EndTurnH
}

func drawCardTooltip(screen *ebiten.Image, c *combat.Combat) {
	mx, my := ebiten.CursorPosition()
	i := HitCard(c, mx, my)
	if i < 0 {
		return
	}
	card := c.Hand[i]
	cx, cy := cardRect(i)

	const (
		tipW = 240
		tipH = 60
		pad  = 8
	)
	tx := cx + (CardW-tipW)/2
	if tx < 4 {
		tx = 4
	}
	if tx+tipW > ScreenW-4 {
		tx = ScreenW - 4 - tipW
	}
	ty := cy - tipH - 8

	vector.DrawFilledRect(screen, float32(tx), float32(ty), tipW, tipH, tooltipBg, true)
	vector.StrokeRect(screen, float32(tx), float32(ty), tipW, tipH, 1, tooltipEdge, true)
	header := fmt.Sprintf("%s %s   (cost %d)", card.Glyph, card.Name, card.Cost)
	drawText(screen, header, tx+pad, ty+pad, faceSmall, white)
	drawText(screen, card.Description, tx+pad, ty+pad+18, faceSmall, white)
	if ok, why := cardPlayable(c, card); !ok {
		drawText(screen, "("+why+")", tx+pad, ty+pad+34, faceSmall, yellow)
	}
}

// energyLine renders "Energy: X/Y" plus the currently-staged cost when the
// player is composing a spell, so they see how much paying for it will leave.
func energyLine(c *combat.Combat) string {
	staged := c.TotalStagedCost()
	if staged == 0 {
		return fmt.Sprintf("Energy: %d/%d", c.Energy, c.MaxEnergy)
	}
	return fmt.Sprintf("Energy: %d/%d  (-%d staged → %d after)", c.Energy, c.MaxEnergy, staged, c.Energy-staged)
}

func drawHUD(screen *ebiten.Image, v RunView) {
	c := v.Combat
	lines := []string{
		fmt.Sprintf("Class: %s", v.Class),
		fmt.Sprintf("Encounter %d / %d", v.EncounterIdx+1, v.TotalEncounters),
		fmt.Sprintf("HP: %d/%d", c.PlayerHP, c.PlayerMaxHP),
		fmt.Sprintf("Armor: %d", c.PlayerArmor),
		energyLine(c),
		fmt.Sprintf("Move: %.0f", c.MovementBudget),
		fmt.Sprintf("Deck: %d  Discard: %d  Total: %d", len(c.Draw), len(c.Discard), v.DeckSize),
	}
	for i, l := range lines {
		drawText(screen, l, 880, 40+i*18, faceSmall, white)
	}
	drawText(screen, "Click card: play   |   Click radar: move   |   E or button: end turn", 40, 8, faceSmall, white)
}

func drawDashBtn(screen *ebiten.Image, c *combat.Combat) {
	col := cardBg
	if !c.CanDash() {
		col = cardBgDim
	}
	vector.DrawFilledRect(screen, DashBtnX, DashBtnY, DashBtnW, DashBtnH, col, true)
	drawText(screen, "DASH (D)", DashBtnX+30, DashBtnY+24, faceSmall, white)
}

// HitDashBtn returns whether (mx, my) is over the Dash button.
func HitDashBtn(mx, my int) bool {
	return mx >= DashBtnX && mx < DashBtnX+DashBtnW &&
		my >= DashBtnY && my < DashBtnY+DashBtnH
}

func drawEndTurn(screen *ebiten.Image, c *combat.Combat) {
	col := cardBg
	if c.Phase != 0 { // not player turn
		col = cardBgDim
	}
	label := "END TURN"
	if len(c.Stage) > 0 {
		if c.StageIsSlow() {
			label = fmt.Sprintf("CAST SLOW (%d)", len(c.Stage))
		} else {
			label = fmt.Sprintf("CAST (%d)", len(c.Stage))
		}
	}
	vector.DrawFilledRect(screen, EndTurnX, EndTurnY, EndTurnW, EndTurnH, col, true)
	centerText(screen, label, EndTurnX, EndTurnY+20, EndTurnW, faceBody, white)
}

func drawPhaseBanner(screen *ebiten.Image, c *combat.Combat) {
	if c.Phase == combat.PhaseEnemy {
		drawText(screen, "(enemy turn)", 560, 700, faceBody, white)
	}
}

// --- Reward screen ---

func rewardCardRect(i, n int) (int, int) {
	totalW := n*RewardCardW + (n-1)*RewardGap
	startX := (ScreenW - totalW) / 2
	x := startX + i*(RewardCardW+RewardGap)
	return x, RewardTopY
}

func skipBtnRect() (int, int) {
	x := (ScreenW - SkipBtnW) / 2
	y := RewardTopY + RewardCardH + 30
	return x, y
}

func drawRewardOverlay(screen *ebiten.Image, v RunView) {
	// Dim backdrop.
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 180}, true)
	drawText(screen, "Choose a rune to add to your deck.", (ScreenW/2)-130, RewardTopY-40, faceBody, white)

	mx, my := ebiten.CursorPosition()
	for i, card := range v.Rewards {
		x, y := rewardCardRect(i, len(v.Rewards))
		bg := cardBg
		hovered := mx >= x && mx < x+RewardCardW && my >= y && my < y+RewardCardH
		if hovered {
			bg = cardBgHi
		}
		border := cardBorderCol
		if card.Slow {
			border = slowCol
		}
		drawCardFrame(screen, card, x, y, RewardCardW, RewardCardH, bg, border, cardKindReward)
	}

	sx, sy := skipBtnRect()
	bg := cardBg
	if mx >= sx && mx < sx+SkipBtnW && my >= sy && my < sy+SkipBtnH {
		bg = cardBgHi
	}
	vector.DrawFilledRect(screen, float32(sx), float32(sy), SkipBtnW, SkipBtnH, bg, true)
	centerText(screen, "SKIP — take nothing", sx, sy+14, SkipBtnW, faceBody, white)
}


// HitReward returns the index of the reward card at (mx,my), or -1.
func HitReward(rewards []runes.Card, mx, my int) int {
	for i := range rewards {
		x, y := rewardCardRect(i, len(rewards))
		if mx >= x && mx < x+RewardCardW && my >= y && my < y+RewardCardH {
			return i
		}
	}
	return -1
}

func HitSkipReward(mx, my int) bool {
	x, y := skipBtnRect()
	return mx >= x && mx < x+SkipBtnW && my >= y && my < y+SkipBtnH
}

// --- Class select screen ---

const (
	classCardW = 320
	classCardH = 400
	classGap   = 40
	classTopY  = 180
)

type classOption struct {
	class       runes.Class
	title       string
	description []string
}

func classOptions() []classOption {
	return []classOption{
		{
			class: runes.ClassElementalist,
			title: "Elementalist",
			description: []string{
				"Exploits the type system. Match damage type to enemy weakness for 1.5x damage.",
				"",
				"Defense: Earth Armor (cannot move once cast).",
				"No mobility runes — stand and burn.",
				"",
				"Starter: 4 Fire, 3 Earth Armor, 2 Frost.",
			},
		},
		{
			class: runes.ClassMesmer,
			title: "Mesmer",
			description: []string{
				"Reads enemy intent. Punishes both aggression and passivity.",
				"",
				"Defense: positioning and delay.",
				"Aphyr stalls; Move keeps distance.",
				"",
				"Starter: 2 Aphyr, 3 Isa-aggressive, 3 Isa-passive, 2 Move.",
			},
		},
		{
			class: runes.ClassNecromancer,
			title: "Necromancer",
			description: []string{
				"Summons minions as persistent processes that occupy the radar.",
				"",
				"Defense: minions intercept aggression.",
				"Drain heals; sacrifice burns minion HP for damage.",
				"",
				"Starter: 4 Thurisaz, 3 Maðr, 3 Ár.",
			},
		},
	}
}

func classCardRect(i int) (int, int) {
	opts := classOptions()
	n := len(opts)
	totalW := n*classCardW + (n-1)*classGap
	startX := (ScreenW - totalW) / 2
	x := startX + i*(classCardW+classGap)
	return x, classTopY
}

func drawClassSelect(screen *ebiten.Image) {
	centerText(screen, "NAND2RUNES — choose your class", 0, 90, ScreenW, faceTitle, white)

	const (
		pad      = 18
		textMaxW = float64(classCardW - pad*2)
		lineH    = 20
	)

	mx, my := ebiten.CursorPosition()
	for i, opt := range classOptions() {
		x, y := classCardRect(i)
		bg := cardBg
		if mx >= x && mx < x+classCardW && my >= y && my < y+classCardH {
			bg = cardBgHi
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), classCardW, classCardH, bg, true)
		vector.StrokeRect(screen, float32(x), float32(y), classCardW, classCardH, 1, tooltipEdge, true)
		drawText(screen, opt.title, x+pad, y+pad, faceTitle, white)
		cy := y + 56
		for _, line := range opt.description {
			if line == "" {
				cy += lineH / 2
				continue
			}
			cy = drawWrapped(screen, line, x+pad, cy, textMaxW, lineH, faceSmall, white)
		}
	}
}

// HitClassPick returns the class at (mx, my) if any.
func HitClassPick(mx, my int) (runes.Class, bool) {
	for i, opt := range classOptions() {
		x, y := classCardRect(i)
		if mx >= x && mx < x+classCardW && my >= y && my < y+classCardH {
			return opt.class, true
		}
	}
	return 0, false
}

// --- End-of-run overlay ---

func drawEndOverlay(screen *ebiten.Image, v RunView, won bool) {
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 200}, true)
	msg := "RUN COMPLETE"
	if !won {
		msg = "RUN FAILED"
	}
	centerText(screen, msg, 0, ScreenH/2-24, ScreenW, faceTitle, white)
	centerText(screen, fmt.Sprintf("Final HP: %d/%d   Deck size: %d", v.PlayerHP, v.MaxHP, v.DeckSize), 0, ScreenH/2+10, ScreenW, faceBody, white)
	centerText(screen, "Close the window to exit.", 0, ScreenH/2+38, ScreenW, faceSmall, white)
}
