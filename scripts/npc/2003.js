var chat = "Ask me whatever you like! Remember, you'll pick up most of this information just completing quests on Maple Island. #b";
var options = ["How do I move?", "How do I take down the monsters?", "How can I pick up an item?", "What happens when I die?", "When can I choose a job?", "Tell me more about this island!", "What should I do to become a Warrior?", "What should I do to become a Bowman?", "What should I do to become a Thief?", "What should I do to become a Magician?", "What should I do to become a pirate?", "How do I raise the character stats? (S)", "How do I check the items that I just picked up?", "How do I equip an item?", "How do I check out the items that I'm wearing?", "What are skills? (K)", "How do I get to Victoria Island?", "What are mesos?"];

for (var i = 0; i < options.length; i++) {
    chat += "\r\n#L" + i + "#" + options[i] + "#l";
}
npc.sendSelection(chat);
var selection = npc.selection();

var Text = [
    "Alright this is how you move. Use #bleft, right arrow#k to move around the flatland and slanted roads, and press #bAlt#k to jump. A select number of shoes improve your speed and jumping abilities.\r\n\r\n#fUI/DialogImage.img/Help/0#",
    "Here's how to take down a monster. Every monster possesses an HP of its own and you'll take them down by attacking with either a weapon or through spells. Of course the stronger they are, the harder it is to take them down.\r\n\r\n#fUI/DialogImage.img/Help/1#",
    "This is how you gather up an item. Once you take down a monster, an item will be dropped to the ground. When that happens, stand in front of the item and press #bZ#k or #b0 on the NumPad#k to acquire the item.\r\n\r\n#fUI/DialogImage.img/Help/2#",
    "Curious to find out what happens when you die? You'll become a ghost when your HP reaches 0. There will be a tombstone in that place and you won't be able to move, although you still will be able to chat.",
    "When do you get to choose your job? Hahaha, take it easy, my friend. Each job has a requirement set for you to meet. Normally a level between 8 and 10 will do, so work hard.",
    "Want to know about this island? It's called Maple Island and it floats in the air. It's been floating in the sky for a while so the nasty monsters aren't really around. It's a very peaceful island, perfect for beginners!",
    "You want to become a #bWarrior#k? Hmmm, then I suggest you head over to Victoria Island. Head over to a warrior-town called #rPerion#k and see #bDances with Balrog#k. He'll teach you all about becoming a true warrior. Ohh, and one VERY important thing: You'll need to be at least level 10 in order to become a warrior!!",
    "You want to become a #bBowman#k? You'll need to go to Victoria Island to make the job advancement. Head over to a bowman-town called #rHenesys#k and talk to the beautiful #bAthena Pierce#k and learn the in's and out's of being a bowman. Ohh, and one VERY important thing: You'll need to be at least level 10 in order to become a bowman!!",
    "You want to become a #bThief#k? In order to become one, you'll have to head over to Victoria Island. Head over to a thief-town called #rKerning City#k, and on the shadier side of town, you'll see a thief's hideaway. There, you'll meet #bDark Lord#k who'll teach you everything about being a thief. Ohh, and one VERY important thing: You'll need to be at least level 10 in order to become a thief!!",
    "You want to become a #bMagician#k? For you to do that, you'll have to head over to Victoria Island. Head over to a magician-town called #rEllinia#k, and at the very top lies the Magic Library. Inside, you'll meet the head of all wizards, #bGrendel the Really Old#k, who'll teach you everything about becoming a wizard.",
    "Do you wish to become a #bPirate#k?