//Time Setting is in milliseconds
var closeGateTime = 240000; //The time to close the gate
var takeoffTime = 300000; //The time at which takeoff occurs
var landTime = 600000; //The time required to land everyone
var invasionTime = 60000; //The time that balrog invasion starts from takeoffTime between ellinia and orbis
var rogSummonTime = 5000 //The time to spawn the rogs
var rogCheckTime = 10000

var platforms = [101000300, 200000111, 200000121, 220000110] // Ellinia, Orbis (E), Orbis (L) , Ludi

var departureWarps = [
    {src: 101000301, dest: 200090010}, // Ellinia -> Orbis
    {src: 200000112, dest: 200090000}, // Orbis -> Ellinia
    {src: 200000122, dest: 200090100}, // Orbis -> Ludi
    {src: 220000111, dest: 200090110}, // Ludi -> Orbis
]

var arrivalWarps = [
    {src: [200090010, 200090011], dest: 200000100}, // Ellinia (incl. cabin) -> orbis
    {src: [200090000, 200090001], dest: 101000300}, // Orbis (incl. cabin) -> Ellinia
    {src: [200090100], dest: 220000100}, // Orbis -> Ludi
    {src: [200090110], dest: 200000100}, // Ludi -> Orbis
]

var cRogs = [{map: 200090010, x: 485, y: -221}, {map: 200090000, x: -590, y: -221}]
var cRogMaps = [200090010, 200090000]

var shipBoat = 0
var rogBoat = 1

function showBoats(controller, maps, show, type) {
    for (var i = 0; i < maps.length; i++) {
        var instances = controller.fields()[maps[i]].instances()
        for (var j = 0; j < instances.length; j++) {
            instances[j].showBoat(show, type)
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

function init(controller) {
    dock(controller)
}

function dock(controller) {
    showBoats(controller, platforms, true, shipBoat)
    allowTicketSales(controller, platforms, true)

    controller.schedule("closeGate", closeGateTime)
    controller.schedule("takeoff", takeoffTime)
}

function closeGate(controller) {
    allowTicketSales(controller, platforms, false)
}

function takeoff(controller) {
    showBoats(controller, platforms, false, shipBoat)
    
    for (var i = 0; i < departureWarps.length; i++) {
        movePlayers(controller, departureWarps[i].src, departureWarps[i].dest)
    }

    controller.schedule("invasion", invasionTime)
    controller.schedule("land", landTime)
}

function invasion(controller) {
    chance = Math.random()

    if (chance <= 0.5) {
        return
    }

    controller.log("cRog boat invasion started")
    showBoats(controller, cRogMaps, true, rogBoat)

    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()
        for (var j = 0; j < instances.length; j++) {
            instances[j].changeBgm("Bgm04/ArabPirate")
        }
    }

    controller.schedule("summonRogs", rogSummonTime)
}

function summonRogs(controller) {
    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()
        var pos = controller.createPos(cRogs[i].x, cRogs[i].y)

        for (var j = 0; j < instances.length; j++) {
            instances[j].lifePool().spawnMobFromID(8150000, pos, false, true, true)
            instances[j].lifePool().spawnMobFromID(8150000, pos, false, true, true)
        }
    }

    controller.schedule("checkRogs", rogCheckTime)
}

function checkRogs(controller) {
    var scheduled = false
    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()

        for (var j = 0; j < instances.length; j++) {
            if (instances[j].lifePool().mobCount() > 0) {
                if (!scheduled) {
                    controller.schedule("checkRogs", rogCheckTime)
                    scheduled = true
                }
            } else {
                instances[j].showBoat(false, rogBoat)
            }
        }
    }
}

function land(controller) {
    for (var i = 0; i < arrivalWarps.length; i++) {
        for (var j = 0; j < arrivalWarps[i].src.length; j++) {
            movePlayers(controller, arrivalWarps[i].src[j], arrivalWarps[i].dest)
        }
    }

    showBoats(controller, cRogMaps, false, shipBoat)
    showBoats(controller, cRogMaps, false, rogBoat)

    for (var i = 0; i < cRogs.length; i++) {
        var instances = controller.fields()[cRogs[i].map].instances()

        for (var j = 0; j < instances.length; j++) {
            instances[j].lifePool().eraseMobs()
            instances[j].dropPool().eraseDrops()
        }
    }

    dock(controller)
}