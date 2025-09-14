npc.sendSelection("Hello! Having fun exploring Maple World? \r\n#L0##eDelete Un-droppable items.#l\r\n#L1#End conversation#l");
var sel = npc.selection();

if (sel === 0) {
    npc.sendOk("Delete Un-droppable items – understood and executed!");
}