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

## Connecting To Your Running Game Server (running in a docker container)
<img height="43px" src="https://d29fhpw069ctt2.cloudfront.net/icon/image/38771/preview.svg"/>

## Administrating Your Server
The server sends auditing information to an eleastic search instance running in a docker container. Kibana can be used to view the data.

GM Commands, chat, login/logout events, server transitions, party join/leave, trade transactions, damage received/inflicted, skill used, stat distribution  etc are logged into elastic search for auditing purposes to be used to find hackers and detect any harassmment etc

#### GM Commands - prefix of ***!***
* packet - Send packet to GM client (used for development)
* warp - warp to map
* notice - send notice message to channel
* dialogue - send dialogue box message to channel
* job - change job
* level - change level
* spawn - spawn mob at character location
* killmobs - kill all mobs on map
* exp - give exp
* mobrate - modify mob rate of server
* exprate - modify exp rate
* mesorate - modify meso rate
* droprate - modify drop rate
* header - change channel header message

Check command/handlers.go for parameters

## Roadmap
* save changes to database
* improve monster spawn system
* monster/player drop items
* inventory management
* parties
* guilds
* npc scripting system

<div>Valhalla Logo made with <a href="https://
www.designevo.com/" title="Free Online Logo Maker">DesignEvo</a></div>