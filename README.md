<p align="center">
  <img src="https://i.imgur.com/mo4tfJF.png"/>
</p>

# Valhalla
Golang v28 maplestory server written in golang. 

## Compiling Instructions
Install [docker](https://docs.docker.com/install/) & [docker-compose](https://docs.docker.com/compose/install/). Thats it!

## Starting the Server

In order to run the channel server you are required to convert a v28 Data.wz (not provided) to Data.nx ([NX](https://nxformat.github.io/) file format). Note: Make sure to disable re-ordering in for items when performing the conversion. This is needed as Wizet did not match index with id. If the order of items is re-ordered based on id then you will end up with mismatched portals.

docker-compose up will start the server components in docker containers (note: database data is stored as docker volume)

***note: atm the game servers are not dockerised***

## Connecting To Your Running Containers
tbc
## Roadmap
* save changes to database
* improve monster spawn system

<div>Valhalla Logo made with <a href="https://
www.designevo.com/" title="Free Online Logo Maker">DesignEvo</a></div>