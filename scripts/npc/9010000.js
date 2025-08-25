npc.sendSelection("Hello! Having fun exploring Maple World? \r\n#L0##eDelete Un-droppable items.#l\r\n#L1#End conversation#l")
var sel = npc.selection()

if (sel == 0) {
    npc.sendOk("Delete Un-droppable items feature is not yet implemented.")
} else if (sel == 1) {
    npc.sendOk("Goodbye!")
}

// Generate by kimi-k2-instruct