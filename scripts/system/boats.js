//Time Setting is in milliseconds
var closeGateTime = 240000; //The time to close the gate
var takeoffTime = 300000; //The time at which takeoff occurs
var landTime = 600000; //The time required to land everyone

var invasionTime = 60000; //The time that spawn balrog from takeoffTime between ellinia and orbis

var stations = [101000300]

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

function run(controller) {
    controller.log("showBoats")

    // remove balrog ship, change bgm back to normal, kill any existing crogs

    showBoats(controller, stations, true)
    allowTicketSales(controller, stations, true)

    controller.schedule("closeGate", closeGateTime)
    controller.schedule("takeoff", takeoffTime)
}

function closeGate(controller) {
    controller.log("closeGate")
    allowTicketSales(controller, stations, false)
}

function takeoff(controller) {
    controller.log("takeoff")
    showBoats(controller, stations, false)
    
    // move characters from boat waiting maps to boat flying map in all instances

    controller.schedule("invasion", invasionTime)
    controller.schedule("land", landTime)
}

function invasion(controller) {
    controller.log("invasion")
    chance = Math.random()
    controller.log(chance)
    if (chance <= 0.5) {
        controller.log("not happening")
        return
    }

    controller.log("happening")
    // 40% change to spawn
    // spawn 8150000 at 485, -221
    // spawn 8150000 at -590, -221
    // show boat
    // change map bgm (Bgm04/ArabPirate)
}

function land(controller) {
    controller.log("land")
    // move characters from boat to stations in all instances

    run(controller)
}