// Ms. Tan — Henesys Skin-Care (REG coupon, preview + apply)
const couponSkin = 5153000

// Intro
npc.sendBackNext(
    "Welcome to Henesys Skin-Care! For just one teeny-weeny #b#t" + couponSkin + "##k, I can make your skin supple and glow-y, like mine! Trust me, you don't want to miss my facials.",
    false, true
)

// Supported skin tones for this version
var skin = [0, 1, 2, 3, 4]

// Show preview/selection UI (expand the array for variadic API)
npc.sendAvatar.apply(npc,
    ["With our specialized machine, you can see your skin after the treatment in advance. Which would you like?"].concat(skin)
)
var sel = npc.selection()

// Validate selection
if (sel < 0 || sel >= skin.length) {
    npc.sendOk("Changed your mind? That's fine. Come back any time.")
} else if (plr.itemCount(couponSkin) >= 1) {
    plr.removeItemsByID(couponSkin, 1)
    plr.setSkinColor(skin[sel])
    npc.sendOk("Enjoy your new and improved skin color!")
} else {
    npc.sendOk("Um... you don't have the skin-care coupon you need to receive the treatment. Sorry, but we can't do it for you...")
}