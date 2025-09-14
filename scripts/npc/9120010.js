var item = [4000064, 4000065, 4000066, 4000075, 4000077, 4000089, 4000090, 4000091, 4000092, 4000093, 4000094];
var Prizes = [2022019, 2022019, 2022019, 2022022, 2022022, 2022022, 2022026, 2001002, 2022000];
var num = [5, 10, 15, 5, 10, 15, 15, 15, 20];

var txt = ["Hmmm...I see some dents here and there. And what's this scratch here? Did you run into a wild cat? Honestly, this is below the standard I've come to expect from warriors such as yourself. As always, I will reward you with an item of similar quality. Here, I'll give you Kinoko Ramen(pig head).", 
	"Hmmm...if not for this minor scratch...sigh. I'm afraid I can only deem this a standard-quality item. Well, here's Kinoko Ramen(pig head) for you.", 
	"Hmmm...if not for this minor scratch...sigh. I'm afraid I can only deem this a standard-quality item. Well, here's Kinoko Ramen(pig head) for you.", 
	"Hmmm...I see some dents here and there. And what's this scratch here? Did you run into a wild cat? Honestly, this is below the standard I've come to expect from warriors such as yourself. As always, I will reward you with an item of similar quality. Here, I'll give you Fish Cake(dish).", 
	"Hmmm...if not for this minor scratch...sigh. I'm afraid I can only deem this a standard-quality item. Well, here's Fish Cake(dish) for you.", 
	"Hmmm...if not for this minor scratch...sigh. I'm afraid I can only deem this a standard-quality item. Well, here's Fish Cake(dish) for you.", 
	"Ohh... I like this. Yes! This is definitely something that cannot be easily obtained. No doubt this is going to be part of my collection. I can't believe that you not only found this, but also gathered up mass quantities! Something as awesome as this deserves a similarly great reward like Yakisoba. It's okay, please go ahead and receive it!", 
	"Ohh... I like this. Yes! This is definitely something that cannot be easily obtained. No doubt this is going to be part of my collection. I can't believe that you not only found this, but also gathered up mass quantities! Something as awesome as this deserves a similarly great reward like Very Special Sundae. It's okay, please go ahead and receive it!", 
	"Hmmm...if not for this minor scratch...sigh. I'm afraid I can only deem this a standard-quality item. Well, here's Pure Water for you."];

if (!npc.sendYesNo("If you're looking for someone who can pinpoint the characteristics of various items, you're looking at him now. Would you like to hear my story?")) {
    npc.sendOk("Really? Let me know if you ever change your mind. I'll be waiting!");
}

var chat = "The items I'm looking for are...ugh, just too many to mention. But if you gather up 100 of the same items, then I may trade it for something similar. I can understand being a little wary, but don't worry--I'll keep my end of the deal. Now, shall we trade? #b"
for (var i = 0; i < item.length; i++)
    chat += "\r\n#L" + i + "##v" + item[i] + "##z" + item[i] + "##l";
npc.sendSelection(chat);
var selection = npc.selection();

if (plr.itemCount(item[selection]) < 100) {
    npc.sendNext("Hey, what do you think you're doing? Go lie to some other unassuming fellow. Not me!");
}

// Simulate free slots check by attempting exchange (or rely on giveItem result)
var rand = Math.floor(Math.random() * Prizes.length);

if (!plr.giveItem(Prizes[rand], num[rand])) {
    npc.sendNext("What? I can't give you the reward if your Equip, Use, or Etc window in your Item Inventory is full. Please make room and I will gladly give you what you came for.");
}

plr.removeItemsByID(item[selection], 100);
npc.sendNext(txt[rand]);