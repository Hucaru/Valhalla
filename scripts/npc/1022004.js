// Mr. Smith – stateless, top-level async flow

var selType, selItem, itemCode, matArr, matQtyArr, costEach, qty;

npc.sendSelection(
    "Um... Hi, I'm Mr. Thunder's apprentice. He's getting up there in age, so he handles most of the heavy-duty work while I handle some of the lighter jobs. What can I do for you?#b"
    + "\r\n#L0# Make a glove#l"
    + "\r\n#L1# Upgrade a glove#l"
    + "\r\n#L2# Create materials#l"
);
selType = npc.selection();

switch (selType) {
    // ── make glove ----------------------------------------------------------
    case 0: {
        npc.sendSelection(
            "Okay, so which glove do you want me to make?#b"
            + "\r\n#L0# Juno#k - Warrior Lv. 10#l"
            + "\r\n#L1# Steel Fingerless Gloves#k - Warrior Lv. 15#l"
            + "\r\n#L2# Venon#k - Warrior Lv. 20#l"
            + "\r\n#L3# White Fingerless Gloves#k - Warrior Lv. 25#l"
            + "\r\n#L4# Bronze Missel#k - Warrior Lv. 30#l"
            + "\r\n#L5# Steel Briggon#k - Warrior Lv. 35#l"
            + "\r\n#L6# Iron Knuckle#k - Warrior Lv. 40#l"
            + "\r\n#L7# Steel Brist#k - Warrior Lv. 50#l"
            + "\r\n#L8# Bronze Clench#k - Warrior Lv. 60#l"
        );
        selItem = npc.selection();

        const gloves = [1082003,1082000,1082004,1082001,1082007,1082008,1082023,1082009,1082059];
        const mats   = [
            [4000021,4011001],          // Juno
            [4011001],                  // Steel Fingerless
            [4000021,4011000],          // Venon
            [4011001],                  // White Fingerless
            [4011000,4011001,4003000],  // Bronze Missel
            [4000021,4011001,4003000],  // Steel Briggon
            [4000021,4011001,4003000],  // Iron Knuckle
            [4011001,4021007,4000030,4003000],
            [4011007,4011000,4011006,4000030,4003000]
        ];
        const counts = [
            [15,1], [2],            // 0
            [40,2], [2],            // 2
            [3,2,15], [30,4,15],    // 4,5
            [50,5,40],              // 6
            [3,2,30,45],            // 7
            [1,8,2,50,50]           // 8
        ];
        const prices = [1000,2000,5000,10000,20000,30000,40000,50000,70000];

        itemCode  = gloves[selItem];
        matArr    = mats[selItem];
        matQtyArr = counts[selItem];
        costEach  = prices[selItem];
        qty       = 1;
        break;
    }

    // ── upgrade glove -------------------------------------------------------
    case 1: {
        npc.sendSelection(
            "Upgrade a glove? That shouldn't be too difficult. Which did you have in mind?#b"
            + "\r\n#L0# Steel Missel#k - Warrior Lv. 30#l"
            + "\r\n#L1# Orihalcon Missel#k - Warrior Lv. 30#l"
            + "\r\n#L2# Yellow Briggon#k - Warrior Lv. 35#l"
            + "\r\n#L3# Dark Briggon#k - Warrior Lv. 35#l"
            + "\r\n#L4# Adamantium Knuckle#k - Warrior Lv. 40#l"
            + "\r\n#L5# Dark Knuckle#k - Warrior Lv. 40#l"
            + "\r\n#L6# Mithril Brist#k - Warrior Lv. 50#l"
            + "\r\n#L7# Gold Brist#k - Warrior Lv. 50#l"
            + "\r\n#L8# Sapphire Clench#k - Warrior Lv. 60#l"
            + "\r\n#L9# Dark Clench#k - Warrior Lv. 60#l"
        );
        selItem = npc.selection();

        const upGloves = [1082005,1082006,1082035,1082036,1082024,1082025,1082010,1082011,1082060,1082061];
        const upMats   = [
            [1082007,4011001],
            [1082007,4011005],
            [1082008,4021006],
            [1082008,4021008],
            [1082023,4011003],
            [1082023,4021008],
            [1082009,4011002],
            [1082009,4011006],
            [1082059,4011002,4021005],
            [1082059,4021007,4021008]
        ];
        const upCounts = [
            [1,1],[1,2],
            [1,3],[1,1],
            [1,4],[1,2],
            [1,5],[1,4],
            [1,3,5],
            [1,2,2]
        ];
        const upCost   = [20000,25000,30000,40000,45000,50000,55000,60000,70000,80000];

        itemCode  = upGloves[selItem];
        matArr    = upMats[selItem];
        matQtyArr = upCounts[selItem];
        costEach  = upCost[selItem];
        qty       = 1;
        break;
    }

    // ── create materials ----------------------------------------------------
    case 2: {
        npc.sendSelection(
            "Materials? I know of a few materials that I can make for you...#b"
            + "\r\n#L0# Make Processed Wood with Tree Branch#l"
            + "\r\n#L1# Make Processed Wood with Firewood#l"
            + "\r\n#L2# Make Screws (packs of 15)#l"
        );
        selItem = npc.selection();

        const materialItem = [4003001,4003001,4003000];
        const materialMat  = [4000003,4000018,[4011000,4011001]];
        const materialQty  = [10,5,[1,1]];
        const materialP    = [0,0,0];

        itemCode  = materialItem[selItem];
        matArr    = materialMat[selItem];
        matQtyArr = materialQty[selItem];
        costEach  = materialP[selItem];

        let prompt = "So, you want me to make some #t" + itemCode + "#? In that case, how many do you want me to make?";
        qty = npc.askNumber(prompt,1,1,100);
        break;
    }

    default:
        npc.sendOk("No valid route.");
}

// ------------------------------------------------------------------
// Common confirmation
var prompt2 = "You want me to make ";
prompt2 += (qty === 1 ? "a #t" + itemCode + "#?" : qty + " #t" + itemCode + "#?");
prompt2 += " In that case, I'm going to need specific items from you in order to make it. Make sure you have room in your inventory, though!#b";

if (matArr instanceof Array) {
    for (let i=0;i<matArr.length;i++){
        prompt2 += "\r\n#i" + matArr[i] + "# " + (matQtyArr[i]*qty) + " #t" + matArr[i] + "#";
    }
} else {
    prompt2 += "\r\n#i" + matArr + "# " + (matQtyArr*qty) + " #t" + matArr + "#";
}
if (costEach)
    prompt2 += "\r\n" + (costEach*qty) + " meso";

if (!npc.sendYesNo(prompt2)) {
    npc.sendOk("Talk to me again anytime.");
}

// ------------------------------------------------------------------
// Validation & creation
let ok = true;
let totalCost = costEach * qty;

if (plr.mesos() < totalCost) {
    npc.sendOk("I may still be an apprentice, but I do need to earn a living.");
}

function invSafe(itemId, need){
    return plr.itemCount(itemId) >= need;
}

if (matArr instanceof Array){
    for (let i=0;i<matArr.length;i++){
        if (!invSafe(matArr[i], matQtyArr[i]*qty)){
            ok=false;
            break;
        }
    }
} else {
    if (!invSafe(matArr, matQtyArr*qty)) ok=false;
}

if (!ok){
    npc.sendOk("I'm still an apprentice, I don't know if I can substitute other items in yet... Can you please bring what the recipe calls for?");
}

// ---- deduct mats & mesos, grant product ----
if (matArr instanceof Array){
    for (let i=0;i<matArr.length;i++)
        plr.removeItemsByID(matArr[i], matQtyArr[i]*qty);
} else {
    plr.removeItemsByID(matArr, matQtyAr*qty);
}

plr.takeMesos(totalCost);

if (itemCode === 4003000)           // screws pack
    plr.giveItem(4003000, 15 * qty);
else
    plr.giveItem(itemCode, qty);

npc.sendOk("Did that come out right? Come by me again if you have anything for me to practice on.");