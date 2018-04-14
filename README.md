<p align="center">
  <img src="https://i.imgur.com/mo4tfJF.png"/>
</p>

## Compiling Instructions
Install [docker](https://docs.docker.com/install/) & [docker-compose](https://docs.docker.com/compose/install/). Thats it! 

## Starting the Server

In order to run the channel server you are required to convert a v28 Data.wz (not provided) to Data.nx ([NX](https://nxformat.github.io/) file format). Note: Make sure to disable re-ordering for items when performing the conversion. This is needed as Wizet did not match index with id. If the order of items is re-ordered based on id then you will end up with mismatched portals.

* docker-compose up -d && docker-compose logs -f will start the server components in docker containers
* ctrl+c will stop displaying logs but the servers will be running in the background still
* To stop everything run docker-compose down
* To restart a specific container run docker-compose restart \<name e.g. login-server\>, if a container crashes it will auto restart
* To stop/start a single container run docker-compose stop/start \<name e.g. login-server\>
* To rebuild and start a container incase of source updates run docker-compose build && docker-compose up -d --no-deps \<name e.g. channel-server\>, it will say that the login-server is being rebuilt, this is because all the servers run off of the same base image

***note: curently the logins server tells the client the channel server at a fixed ip address. change this before docker-compose build***

***note: database data is stored as docker volume***

***note: make sure to configure the services for your ip addresses and ports in the docker-compose.yaml file***

The following is an example of what the docker logs should look like:
![](https://i.imgur.com/Lqh0Ln7.png)

## Connecting To Your Running Game Server (running in a docker container)
<img height="43px" src="https://d29fhpw069ctt2.cloudfront.net/icon/image/38771/preview.svg"/>

## Administrating Your Server
The server sends auditing information to an eleastic search instance running in a docker container. Kibana can be used to view the data.

GM Commands, chat, login/logout events, server transitions, party join/leave, trade transactions, damage received/inflicted, skill used, stat distribution  etc are logged.

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
* redo wizet data loading to be nx & wz agnostic

<div>Valhalla Logo made with <a href="https://
www.designevo.com/" title="Free Online Logo Maker">DesignEvo</a></div>