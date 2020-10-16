//Time Setting is in milliseconds
var closeGateTime = 240000; //The time to close the gate
var takeoffTime = 300000; //The time at which takeoff occurs
var landTime = 600000; //The time required to land everyone
var invasionTime = 60000; //The time that spawn balrog from takeoffTime between ellinia and orbis

var stations = [101000300, 200000111, 200000121, 220000110]

var departureWarps = [
    {src: 101000301, dest: 200090010}, // Ellinia -> Orbis
    {src: 200000112, dest: 200090000}, // Orbis -> Ellinia
    // {src: 200000122, dest: }, // Orbis -> Ludi
    // {src: 220000111, dest: }, // Ludi -> Orbis
]

var arrivalWarps = [
    {src: [200090010, 200090011], dest: 200000100}, // Ellinia (incl. cabin) -> orbis
    {src: [200090000, 200090001], dest: 101000300}, // Orbis (incl. cabin) -> Ellinia
]

var cRogs = [{map: 200090010, x: 485, y: -221}, {map: 200090000, x: -590, y: -221}]
var cRogMaps = [200090010, 200090000]

function showBoats(controller, maps, show) {
    for (var i = 0; i < maps.length; i++) {
        var instances = controller.fields()[maps[i]].instances()
        for (var j = 0; j < instances.length; j++) {
            instances[j].showBoat(show)
        }
    }
}

function allowTicketSales(controller, maps, allow) {
    for (var i = 0; i < maps.length; i++) {
        var instances = controller.fields()[maps[i]].instances()
        for (var j = 0; j < instances.length; j++) {
            instances[j].properties()["canSellTickets"] = allow
        }
    }
}

function movePlayers(controller, source, destination) {
    var sourceInstances = controller.fields()[source].instances()

    for (var i = 0; i < sourceInstances.length; i++) {
        var players = controller.fields()[source].instances()[i].players()

        for (var j = 0; j < players.length; j++) {
            controller.warpPlayerToPortal(players[j], destination, 0)
        }
    }
}

function run(controller) {
    showBoats(controller, stations, true)
    allowTicketSales(controller, stations, true)

    controller.schedule("closeGate", closeGateTime)
    controller.schedule("takeoff", takeoffTime)
}

function closeGate(controller) {
    allowTicketSales(controller, stations, false)
}

function takeoff(controller) {
    showBoats(controller, stations, false)
    
    for (var i = 0; i < departureWarps.length; i++) {
        movePlayers(controller, departureWarps[i].src, departureWarps[i].dest)
    }

    controller.schedule("invasion", invasionTime)
    controller.schedule("land", landTime)
}

function invasion(controller) {
    chance = Math.random()
    controller.log(chance)
    if (chance <= 0.5) {
        return
    }

    controller.log("cRog boat invasion started")
    // 40% change to spawn
    // spawn 8150000 at 485, -221
    // spawn 8150000 at -590, -221
    // show boat
    showBoats(controller, cRogMaps, true)

    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()
        var pos = controller.createPos(cRogs[i].x, cRogs[i].y)

        for (var j = 0; j < instances.length; j++) {
            instances[j].lifePool().spawnMobFromID(8150000, pos, false, true, true)
            instances[j].lifePool().spawnMobFromID(8150000, pos, false, true, true)
            instances[j].changeBgm("Bgm04/ArabPirate")
        }
    }
}

function land(controller) {
    for (var i = 0; i < arrivalWarps.length; i++) {
        for (var j = 0; j < arrivalWarps[i].src.length; j++) {
            movePlayers(controller, arrivalWarps[i].src[j], arrivalWarps[i].dest)
        }
    }

    showBoats(controller, cRogMaps, false)

    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()

        for (var j = 0; j < instances.length; j++) {
            instances[j].lifePool().eraseMobs()
            instances[j].dropPool().eraseDrops()
        }
    }

    run(controller)
}