var questid = [3615, 3616, 3617, 3618, 3920, 3630, 3633, 3639];
var questitem = [4031235, 4031236, 4031237, 4031238, 4031591, 4031270, 4031280, 4031298];

var num = 0;
var books = "";

for (var i = 0; i < questid.length; i++) {
    if (plr.getQuestStatus(questid[i]) > 1) {
        books += "\r\n#v" + questitem[i] + "##t" + questitem[i] + "#";
        num += 1;
    }
}

if (num < 1) {
    npc.sendOk("#b#h0##k has not returned a single storybook yet.");
} else {
    npc.sendBackNext("Let's see.. #b#h ##k have returned a total of #b" + num + "#k books. The list of returned books is as follows: \r\n#b" + books, false, true);
    npc.sendBackNext("The library is settling down now thanks chiefly to you, #b#h0##k's immense help. If the story gets mixed up once again, then I'll be counting on you to fix it once more.", true, false);
}

// Generate by kimi-k2-instruct