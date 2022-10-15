package channel

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type foothold struct {
	id               int16
	x1, y1, x2, y2   int16
	prev, next       int
	centreX, centreY int16
}

func createFoothold(id, x1, y1, x2, y2 int16, prev, next int) foothold {
	return foothold{id: id, x1: x1, y1: y1, x2: x2, y2: y2, prev: prev, next: next, centreX: (x2 + x1) / 2, centreY: (y2 + y1) / 2}
}

// Slope if y1 == y2
func (data foothold) slope() bool {
	return data.y1 != data.y2
}

// Wall if x1 == x2
func (data foothold) wall() bool {
	return data.x1 == data.x2
}

func withinX(check, x1, x2 int16) bool {
	if check >= x1 && check <= x2 {
		return true
	}

	return false
}

func crossProduct(x, x1, x2, y, y1, y2 int16) float64 {
	/*
		cp = |a||b|sin(theta)

		whend dealing with vectors it can be calculated as:

		cp.x = a.y * b.z - a.z * b.y
		cp.y = a.z * b.x - a.x * b.z
		cp.z = a.x * b.y - a.y * b.x

		working in 2d therefore z is zero meaning cp.x & cp.y do not need to be calculated

		since x & y component of cp vector are zero cp.z is the vector magnitude

		|cp| / |a|.|b| = sin(theta)

		if theta lies between 0 and pi crossing pi/2 then it is above the line resulting in a positive value
		if theta lies between 0 and pi crossing 3pi/2 then it is below the line resulting in a negative value
	*/
	return float64(x-x1)*float64(y2-y1) - float64(y-y1)*float64(x2-x1) // 0 is on the line, > 0 is above, < 0 is below
}

func (data foothold) above(p pos, ignoreX bool) bool {
	if !withinX(p.x, data.x1, data.x2) && !ignoreX {
		return false
	}

	return crossProduct(p.x, data.x1, data.x2, p.y, data.y1, data.y2) >= 0
}

func (data foothold) findPos(p pos) pos {
	if !data.slope() {
		return newPos(p.x, data.y1, data.id)
	}

	/*
		Equation derived for two collinear points as follows:
		P1 + k(P1 - P2) = R
		x1 + k(x1 - x2) = rx
		k = (rx - x1) / (x1 - x2)

		y1 + k(y1 - y2) = ry

		ry = y1 + ((rx - x1) / (x1 - x2)) * (y1 - y2)

		pre-calculating y1 - y2 and x1 - x2 might yield perf increases (extremely minor)
	*/

	newY := data.y1 + int16((float64(p.x-data.x1)/float64(data.x1-data.x2))*float64(data.y1-data.y2))

	return newPos(p.x, newY, data.id)
}

func (data foothold) distanceFromPosSquare(point pos) (int16, int16, int16) {
	deltaX := point.x - data.centreX
	deltaY := point.y - data.centreY

	clampX := data.x1 + 30
	clampY := data.y1

	if deltaX > 0 {
		clampX = data.x2 - 30
		clampY = data.y2
	}

	return (deltaX * deltaX) + (deltaY * deltaY), clampX, clampY
}

// Histogram of foothold data, aims to reduce the amount of footholds that are iterated and compared against one another
// compared to iterating over the slice of all footholds
type fhHistogram struct {
	footholds []foothold
	binSize   int
	minX      int16
	bins      [][]*foothold
}

func createFootholdHistogram(footholds []foothold) fhHistogram {
	var minX int16
	var maxX int16

	for _, v := range footholds {
		if v.x1 == v.x2 { // Ignore walls as it scuffs the offsets for some narrow maps
			continue
		}

		if v.x1 < minX {
			minX = v.x1
		}

		if v.x2 > maxX {
			maxX = v.x2
		}
	}

	delta := maxX - minX
	binSize := int(math.Ceil(float64(delta) / float64(len(footholds))))
	binCount := int(math.Ceil(float64(delta) / float64(binSize)))
	bins := make([][]*foothold, binCount+1)

	result := fhHistogram{footholds: footholds, binSize: binSize, minX: minX, bins: bins}

	for i, v := range result.footholds {
		if v.x1 == v.x2 { // Ignore walls
			continue
		}

		first := result.calculateBinIndex(v.x1)
		last := result.calculateBinIndex(v.x2)

		for j := first; j <= last; j++ {
			result.bins[j] = append(result.bins[j], &result.footholds[i])
		}
	}

	return result
}

// MarshalJSON interface conformality for debug purposes
func (data fhHistogram) MarshalJSON() ([]byte, error) {
	bins := make([]int, len(data.bins))

	for i := range bins {
		bins[i] = len(data.bins[i])
	}

	return json.Marshal(struct {
		Bins    []int
		MinX    int16
		BinSize int
	}{
		bins,
		data.minX,
		data.binSize,
	})
}

func (data fhHistogram) calculateBinIndex(x int16) int {
	ind := x - data.minX

	if ind > 0 {
		ind = int16(math.Ceil(float64(ind) / float64(data.binSize)))
	} else if ind == 0 {
		ind = 0
	} else {
		ind = -1
	}

	return int(ind)
}

func (data fhHistogram) getFinalPosition(point pos) pos {
	ind := data.calculateBinIndex(point.x)

	if ind < 0 {
		return data.findNearestPoint(0, point)
	} else if ind > len(data.bins)-1 {
		return data.findNearestPoint(len(data.bins)-1, point)
	}

	return data.retrivePosition(ind, point)
}

func (data fhHistogram) retrivePosition(ind int, point pos) pos {
	minimum := point
	set := false

	for _, v := range data.bins[ind] {
		if !v.wall() && v.above(point, false) {
			pos := v.findPos(point)

			if pos.y >= point.y {
				if !set {
					set = true
					minimum = pos
				} else if pos.y < minimum.y {
					minimum = pos
				}
			}
		}
	}

	if !set {
		minimum = data.findNearestPoint(ind, point)
	}

	return minimum
}

func (data fhHistogram) findNearestPoint(ind int, point pos) pos {
	nearest := point

	var dist int16 = math.MaxInt16

	for _, v := range data.bins[ind] {
		if !v.wall() && v.above(point, true) {
			if d, clampX, clampY := v.distanceFromPosSquare(point); d < dist {
				dist = d
				nearest.x = clampX
				nearest.y = clampY
			}
		}
	}

	return nearest
}

type fieldRectangle struct {
	Left, Top, Right, Bottom int64
}

// create from LTRB from:
// x-coordinate of the upper-left corner,
// y-coordinate of the upper-left corner,
// x-coordinate of the lower-right corner,
// y-coordinate of the lower-right corner
func createFromLTRB(left, top, right, bottom int64) fieldRectangle {
	return fieldRectangle{Left: left, Top: top, Right: right, Bottom: bottom}
}

func (data fieldRectangle) inflate(x, y int64) fieldRectangle {
	xDelta := x / 2
	yDelta := y / 2

	return fieldRectangle{Left: data.Left - xDelta,
		Top:    data.Top + yDelta,
		Right:  data.Right + xDelta,
		Bottom: data.Bottom - yDelta,
	}
}

func (data fieldRectangle) empty() bool {
	if data.Left == 0 && data.Top == 0 && data.Right == 0 && data.Bottom == 0 {
		return true
	}

	return false
}

func (data fieldRectangle) width() int64 {
	return int64(math.Abs(float64(data.Left) - float64(data.Right)))
}

func (data fieldRectangle) height() int64 {
	return int64(math.Abs(float64(data.Top) - float64(data.Bottom)))
}

type field struct {
	id        int32
	instances []*fieldInstance
	Data      nx.Map

	deltaX, deltaY float64

	Dispatch chan func()

	vrLimit                        fieldRectangle
	mobCapacityMin, mobCapacityMax int

	footholds []foothold
	fhHist    fhHistogram
}

func (f *field) createInstance(rates *rates) int {
	id := len(f.instances)

	portals := make([]portal, len(f.Data.Portals))
	for i, p := range f.Data.Portals {
		portals[i] = createPortalFromData(p)
		portals[i].id = byte(i)
	}

	inst := &fieldInstance{
		id:          id,
		fieldID:     f.id,
		portals:     portals,
		dispatch:    f.Dispatch,
		town:        f.Data.Town,
		returnMapID: f.Data.ReturnMap,
		timeLimit:   f.Data.TimeLimit,
		properties:  make(map[string]interface{}),
		fhHist:      f.fhHist,
	}

	inst.roomPool = createNewRoomPool(inst)
	inst.dropPool = createNewDropPool(inst, rates)
	inst.lifePool = creatNewLifePool(inst, f.Data.NPCs, f.Data.Mobs, f.mobCapacityMin, f.mobCapacityMax)
	inst.lifePool.setDropPool(&inst.dropPool)

	f.instances = append(f.instances, inst)

	return id
}

func (f *field) formatFootholds() {
	f.footholds = make([]foothold, len(f.Data.Footholds))

	for i, v := range f.Data.Footholds {
		f.footholds[i] = createFoothold(v.ID, v.X1, v.Y1, v.X2, v.Y2, v.Next, v.Prev)
	}

	f.fhHist = createFootholdHistogram(f.footholds)
}

func (f *field) calculateFieldLimits() {
	vrLimit := createFromLTRB(f.Data.VRLeft, f.Data.VRTop, f.Data.VRRight, f.Data.VRBottom)

	var left int64 = math.MaxInt32
	var top int64 = math.MaxInt32
	var right int64 = math.MinInt32
	var bottom int64 = math.MinInt32

	for _, fh := range f.Data.Footholds {
		if int64(fh.X1) < left {
			left = int64(fh.X1)
		}

		if int64(fh.Y1) < top {
			top = int64(fh.Y1)
		}

		if int64(fh.X2) < left {
			left = int64(fh.X2)
		}

		if int64(fh.Y2) < top {
			top = int64(fh.Y2)
		}

		if int64(fh.X1) > right {
			right = int64(fh.X1)
		}

		if int64(fh.Y1) > bottom {
			bottom = int64(fh.Y1)
		}

		if int64(fh.X2) > right {
			right = int64(fh.X2)
		}

		if int64(fh.Y2) > bottom {
			bottom = int64(fh.Y2)
		}
	}

	if !vrLimit.empty() {
		f.vrLimit = vrLimit
	} else {
		f.vrLimit = createFromLTRB(left, top-300, right, bottom+75)
	}

	left += 30
	top -= 300
	right -= 30
	bottom += 10

	if !vrLimit.empty() {
		if vrLimit.Left+20 < left {
			left = vrLimit.Left + 20
		}

		if vrLimit.Top+65 < top {
			top = vrLimit.Top + 20
		}

		if vrLimit.Right-5 > right {
			right = vrLimit.Right - 5
		}

		if vrLimit.Bottom > bottom {
			bottom = vrLimit.Bottom
		}
	}

	mbr := createFromLTRB(left+10, top-375, right-10, bottom+60)
	mbr = mbr.inflate(10, 10)

	// outofBounds := mbr.Inflate(60, 60)

	var mobX, mobY int64

	if mbr.width() > 800 {
		mobX = mbr.width()
	} else {
		mobX = 800
	}

	if mbr.height()-450 > 600 {
		mobY = mbr.height() - 450
	} else {
		mobY = 600
	}

	var mobCapacityMin int = int(float64(mobX*mobY) * f.Data.MobRate * 0.0000078125)

	if mobCapacityMin < 1 {
		mobCapacityMin = 1
	} else if mobCapacityMin > 40 {
		mobCapacityMin = 40
	}

	mobCapacityMax := mobCapacityMin * 2

	f.mobCapacityMin = mobCapacityMin
	f.mobCapacityMax = mobCapacityMax
}

func (f field) validInstance(instance int) bool {
	if len(f.instances) > instance && instance > -1 {
		return true
	}
	return false
}

func (f *field) deleteInstance(id int) error {
	if f.validInstance(id) {
		if len(f.instances[id].players) > 0 {
			return fmt.Errorf("Cannot delete an instance with players in it")
		}

		f.instances = append(f.instances[:id], f.instances[id+1:]...)

		return nil
	}
	return fmt.Errorf("Invalid instance")
}

func (f *field) getInstance(id int) (*fieldInstance, error) {
	if f.validInstance(id) {
		return f.instances[id], nil
	}

	return nil, fmt.Errorf("Invalid instance id")
}

func (f *field) changePlayerInstance(player *player, id int) error {
	if id == player.inst.id {
		return fmt.Errorf("In specified instance")
	}

	if f.validInstance(id) {
		err := f.instances[player.inst.id].removePlayer(player)

		if err != nil {
			return err
		}

		f.instances[player.inst.id].dropPool.HideDrops(player)

		player.inst = f.instances[id]
		err = f.instances[id].addPlayer(player)

		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Invalid instance id")
}

type portal struct {
	id          byte
	pos         pos
	name        string
	destFieldID int32
	destName    string
	temporary   bool
}

func createPortalFromData(p nx.Portal) portal {
	return portal{id: p.ID,
		pos:         newPos(p.X, p.Y, 0),
		name:        p.Pn,
		destFieldID: p.Tm,
		destName:    p.Tn,
		temporary:   false}
}

type fieldInstance struct {
	id          int
	fieldID     int32
	returnMapID int32
	timeLimit   int64

	lifePool lifePool
	dropPool dropPool
	roomPool roomPool

	portals []portal
	players []*player

	idCounter int32
	town      bool

	dispatch chan func()

	fieldTimer *time.Ticker
	runUpdate  bool

	showBoat   bool
	boatType   byte
	properties map[string]interface{} // this is used to share state between npc and system scripts

	bgm string

	fhHist fhHistogram
}

func (inst fieldInstance) String() string {
	var info string
	info += "field ID: " + strconv.Itoa(int(inst.fieldID)) + ", "
	info += "players(" + strconv.Itoa(len(inst.players)) + "): "

	for _, v := range inst.players {
		info += " " + v.name + "(" + v.pos.String() + ")"
	}

	return info
}

func (inst *fieldInstance) changeBgm(path string) {
	inst.bgm = path
	packetBgmChange(path)
}

func (inst fieldInstance) findController() interface{} {
	for _, v := range inst.players {
		return v
	}

	return nil
}

func (inst *fieldInstance) addPlayer(plr *player) error {
	plr.inst = inst

	for _, other := range inst.players {
		other.send(packetMapPlayerEnter(plr))
		plr.send(packetMapPlayerEnter(other))
	}

	inst.lifePool.addPlayer(plr)
	inst.dropPool.playerShowDrops(plr)
	inst.roomPool.playerShowRooms(plr)

	if inst.showBoat {
		displayBoat(plr, inst.showBoat, inst.boatType)
	}

	inst.players = append(inst.players, plr)

	// For now pools run on all maps forever after first player enters.
	// If this hits perf too much then a set of params for each pool
	// will need to be determined to allow it to stop updating e.g.
	// drop pool, no drops and no players
	// life pool, max number of mobs spawned and no dot attacks in field
	if !inst.runUpdate {
		inst.startFieldTimer()
	}

	if len(inst.bgm) > 0 {
		plr.send(packetBgmChange(inst.bgm))
	}

	return nil
}

func (inst *fieldInstance) removePlayer(plr *player) error {
	index := -1

	for i, v := range inst.players {
		if v.conn == plr.conn {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("player does not exist in instance")
	}

	inst.players = append(inst.players[:index], inst.players[index+1:]...)

	for _, v := range inst.players {
		v.send(packetMapPlayerLeft(plr.id))
		plr.send(packetMapPlayerLeft(v.id))
	}

	inst.lifePool.removePlayer(plr)
	inst.roomPool.removePlayer(plr)

	return nil
}

func (inst fieldInstance) getPlayerFromID(id int32) (*player, error) {
	for i, v := range inst.players {
		if v.id == id {
			return inst.players[i], nil
		}
	}

	return nil, fmt.Errorf("Player not in instance")
}

func (inst fieldInstance) movePlayer(id int32, moveBytes []byte, plr *player) {
	inst.sendExcept(packetPlayerMove(id, moveBytes), plr.conn)
}

func (inst *fieldInstance) nextID() int32 {
	inst.idCounter++
	return inst.idCounter
}

func (inst fieldInstance) send(p mpacket.Packet) error {
	for _, v := range inst.players {
		v.send(p)
	}

	return nil
}

func (inst fieldInstance) sendExcept(p mpacket.Packet, exception mnet.Client) error {
	for _, v := range inst.players {
		if v.conn == exception {
			continue
		}

		v.send(p)
	}

	return nil
}

func (inst fieldInstance) getRandomSpawnPortal() (portal, error) {
	portals := []portal{}

	for _, p := range inst.portals {
		if p.name == "sp" {
			portals = append(portals, p)
		}
	}

	if len(portals) == 0 {
		return portal{}, fmt.Errorf("No spawn portals in map")
	}

	return portals[rand.Intn(len(portals))], nil
}

func (inst fieldInstance) calculateNearestSpawnPortalID(pos pos) (byte, error) {
	var portal portal
	found := true
	err := fmt.Errorf("Portal not found")

	for _, p := range inst.portals {
		if p.name == "sp" && found {
			portal = p
			found = false
			err = nil
		} else if p.name == "sp" {
			delta1 := portal.pos.calcDistanceSquare(pos)
			delta2 := p.pos.calcDistanceSquare(pos)

			if delta2 < delta1 {
				portal = p
			}
		}
	}

	return portal.id, err
}

func (inst fieldInstance) getPortalFromName(name string) (portal, error) {
	for _, p := range inst.portals {
		if p.name == name {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}

func (inst fieldInstance) getPortalFromID(id byte) (portal, error) {
	for _, p := range inst.portals {
		if p.id == id {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}

func (inst *fieldInstance) startFieldTimer() {
	inst.runUpdate = true
	inst.fieldTimer = time.NewTicker(time.Millisecond * 1000) // Is this correct time?

	go func() {
		for t := range inst.fieldTimer.C {
			inst.dispatch <- func() { inst.fieldUpdate(t) }
		}
	}()
}

func (inst *fieldInstance) stopFieldTimer() {
	inst.runUpdate = false
	inst.fieldTimer.Stop()
}

// Responsible for handling the removing of mystic doors, disappearence of loot, ships coming and going
func (inst *fieldInstance) fieldUpdate(t time.Time) {
	inst.lifePool.update(t)
	inst.dropPool.update(t)

	if inst.lifePool.canClose() && inst.dropPool.canClose() {
		inst.stopFieldTimer()
	}
}

func (inst *fieldInstance) calculateFinalDropPos(from pos) pos {
	from.y = from.y - 90 // This distance might need to be configurable depending on drop type? e.g. ludi PQ reward/bonus stage
	return inst.fhHist.getFinalPosition(from)
}

func (inst *fieldInstance) showBoats(show bool, boatType byte) {
	inst.showBoat = show
	inst.boatType = boatType

	for _, v := range inst.players {
		displayBoat(v, show, boatType)
	}
}

func displayBoat(plr *player, show bool, boatType byte) {
	switch boatType {
	case 0: // docked boat in station
		plr.send(packetMapBoat(show))
	case 1: // crog
		plr.send(packetMapShowMovingObject(show))
	default:
		log.Println("Unkown docked boat type:", boatType)
	}
}

func packetMapPlayerEnter(plr *player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterEnterField)
	p.WriteInt32(plr.id)
	p.WriteString(plr.name)

	if true {
		p.WriteString("[Admins]")
		p.WriteInt16(1030) // logo background
		p.WriteByte(3)     // logo bg colour
		p.WriteInt16(4017) // logo
		p.WriteByte(2)     // logo colour
		p.WriteInt32(0)
		p.WriteInt32(0)
	} else {
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
	}

	p.WriteBytes(plr.displayBytes())

	p.WriteInt32(0)           // ?
	p.WriteInt32(0)           // ?
	p.WriteInt32(0)           // ?
	p.WriteInt32(plr.chairID) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(plr.pos.x)
	p.WriteInt16(plr.pos.y)
	p.WriteByte(plr.stance)
	p.WriteInt16(plr.pos.foothold)
	p.WriteInt32(0) // ?

	return p
}

func packetMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func packetMapSpawnMysticDoor(spawnID int32, pos pos, instant bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSpawnDoor)
	p.WriteBool(instant)
	p.WriteInt32(spawnID)
	p.WriteInt16(pos.x)
	p.WriteInt16(pos.y)

	return p
}

func packetMapSpawnTownMysticDoor(dstMap int32, destPos pos) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelTownPortal)
	p.WriteInt32(dstMap)
	p.WriteInt32(dstMap)
	p.WriteInt16(destPos.x)
	p.WriteInt16(destPos.y)

	return p
}

func packetMapRemoveMysticDoor(spawnID int32, instant bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveDoor)
	p.WriteBool(instant)
	p.WriteInt32(spawnID)

	return p
}

func packetMapBoat(show bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBoat)
	if show {
		p.WriteInt16(0x01)
	} else {
		p.WriteInt16(0x02)
	}

	return p
}

func packetMapShowMovingObject(docked bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMovingObj)

	p.WriteByte(0x0a)

	if docked {
		p.WriteByte(4)
	} else {
		p.WriteByte(5)
	}

	return p
}

func packetShowEffect(path string) mpacket.Packet {
	return packetEnvironmentChange(3, path)
}

func packetPlaySound(path string) mpacket.Packet {
	return packetEnvironmentChange(4, path)
}

func packetBgmChange(path string) mpacket.Packet {
	return packetEnvironmentChange(6, path)
}

func packetEnvironmentChange(setting int32, value string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMapEffect)
	p.WriteInt32(setting)
	p.WriteString(value)
	return p
}

// func packetMapPortal(srcMap, dstmap int32, pos pos) mpacket.Packet {
// 	p := mpacket.CreateWithOpcode(0x2d)
// 	p.WriteByte(26)
// 	p.WriteByte(0) // ?
// 	p.WriteInt32(srcMap)
// 	p.WriteInt32(dstmap)
// 	p.WriteInt16(pos.X())
// 	p.WriteInt16(pos.Y())

// 	return p
// }
