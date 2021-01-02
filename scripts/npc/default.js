var state = 0
var styles = [31050, 31040, 31000, 31060, 31090, 31020, 31130, 31120, 31140, 31330, 31010]
var goods = [ [1332020],[1332020, 1],[1332009, 0] ]

function run(npc, player) {
    if (npc.next()) {
        state++
    } else if (npc.back()) {
        state--
    }

    if (state == 2) {
        if (npc.yes()) {
            state = 3
        } else if (npc.no()) {
            state = 4
        }
    } else if (state == 4) {
        if (npc.selection() == 1) {
            state = 0
        } else if (npc.selection() == 2) {
            state = 5
        } else if (npc.selection() == 3) {
            state = 6
        } else if (npc.selection() == 4) {
            state = 7
        } else if (npc.selection() == 5) {
            state = 8
        } else if (npc.selection() == 6) {
            state = 9
        }
    }

    switch(state) {
    case 0:
        npc.sendBackNext("first, npc id: " + npc.id(), false, true)
        break
    case 1:
        npc.sendBackNext("second", true, false)
        break
    case 2:
        npc.sendYesNo("finished")
        break
    case 3:
        npc.sendOK("selection:" + npc.selection() + ", input number:" + npc.inputNumber() + ", input text: " + npc.inputString())
        npc.terminate()
        break
    case 4:
        npc.sendSelection("Select from one of the following:\r\n#L1#Back to start #l\r\n#L2#Styles#l\r\n#L3#Input number#l#L4#Input text#l\r\n#L5#Shop#l\r\n#L6#Warp player#l")
        break
    case 5:
        npc.sendStyles("Select from the following", styles)
        state = 3
        break
    case 6:
        npc.sendInputNumber("Input a number:", 100, 0, 100)
        state = 3
        break
    case 7:
        npc.sendInputText("Input text:", "default", 0, 100)
        state = 3
        break
    case 8:
        npc.sendShop(goods)
        break
    case 9:
        npc.warpPlayer(player, 104000000)
        npc.sendOK("You have been warped")
        npc.terminate()
        break
    default:
        npc.sendOK("state " + state)
        npc.terminate()
    }
    
}