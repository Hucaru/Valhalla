# Valhalla
Golang v28 maplestory server

## TODO:
- Go through and change all packet write, reads to uint variants where appropriate.
- Refactor connection structs into unified base.
- Finish [nx](https://nxformat.github.io/ "Retep998 & angelsl pretty sweet NX File Format [PKG4.1]") data extraction. Once done do optimisation pass.
- Fix channel re-registration that sometimes doesn't work (needs investigation)
- Character delete and select looks at a deleted field. Deletion doesn't delete, just sets flag - allows easy auditing of accounts
- Implement map handler thread
- Implement audit thread (v. low priority)
- REFACTOR EVERYTHING!!!!! (Look out for appends() and see if pre-allocation can be done)

## Usage
### Set-up
### Running

## Some Screenshots

![Alt text](images/server_select.png?raw=true "Server Select")
![Alt text](images/character_select.png?raw=true "Character Select")
![Alt text](images/ingame.png?raw=true "In Game")
