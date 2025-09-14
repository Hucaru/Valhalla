// Stateless conversion
npc.sendNext("Heya, I'm #b#p9000001##k. I came with my friends to enjoy the event, but I can't find them! How about you, will you come with me?")

var sel = npc.sendMenu(
    "Hey... if you aren't busy, do you want to go with me? Looks like my younger sibling will come with others.",
    "What kind of event is it?",
    "Explain the event to me.",
    "Okie! Let's go together!"
)

if (sel === 0) { // What kind of event is it?
    npc.sendBackNext("This event is to celebrate the school vacation! It sucks to be stuck in a room all day, right? So why don't you live vicariously through this exciting vacation event!? Check the event dates on the web site!", false, true)
    npc.sendBackNext("You can obtain various items and mesos from winning in the event! All the event participants will receive trophies while the winners will receive special prizes! Good luck.", true, true)
} else if (sel === 1) { // Explain the event to me
    var gameSel = npc.sendMenu(
        "There are a lot of available events! Wouldn't it be helpful to know the game instructions in advance? Which game do you want to hear the instruction for?",
        "Ola Ola",
        "Physical Fitness Test",
        "Snowball",
        "Coconut Harvest",
        "Speed OX Quiz",
        "Treasure Hunt",
        "Sheep Ranch"
    )
    npc.sendOk(Text[gameSel])
} else if (sel === 2) { // Okie! Let's go together!
    npc.sendNext("You cannot participate in the event if it hasn't started, if you already have Devil's Dice, if you already participated in the event once today, or the participant number cut-off has been reached. Play with me next time!")
}