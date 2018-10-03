<p align="center">
  <img src="https://i.imgur.com/mo4tfJF.png"/>
  <br/>Valhalla Logo made with <a href="https://
www.designevo.com/" title="Free Online Logo Maker">DesignEvo</a>
</p>

## What is this?
This project exists to preserve and archive an early version of the game

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

***note: curently the login server tells the client the channel server at a fixed ip address. change this before docker-compose build***

***note: database data is stored as docker volume***

***note: make sure to configure the services for your ip addresses and ports in the docker-compose.yaml file***

The following is an example of what the docker logs should look like:
![](https://i.imgur.com/Lqh0Ln7.png)

## Roadmap
* Redo login & channel server to handle single events and move away from mutexes
* World server

## NPC chat display info (use this when scripting NPCs)

NPCs are scripted in [anko](https://github.com/mattn/anko)

Taken from [here](http://forum.ragezone.com/f428/add-learning-npcs-start-finish-643364/)
- #b = Blue text.
- #c[itemid]# Shows how many [itemid] the player has in their inventory.
- #d = Purple text.
- #e = Bold text.
- #f[imagelocation]# - Shows an image inside the .wz files.
- #g = Green text.
- #h # - Shows the name of the player.
- #i[itemid]# - Shows a picture of the item.
- #k = Black text.
- #l - Selection close.
- #m[mapid]# - Shows the name of the map.
- #n = Normal text (removes bold).
- #o[mobid]# - Shows the name of the mob.
- #p[npcid]# - Shows the name of the NPC.
- #q[skillid]# - Shows the name of the skill.
- #r = Red text.
- #s[skillid]# - Shows the image of the skill.
- #t[itemid]# - Shows the name of the item.
- #v[itemid]# - Shows a picture of the item.
- #x - Returns "0%" (need more information on this).
- #z[itemid]# - Shows the name of the item.
- #B[%]# - Shows a 'progress' bar.
- #F[imagelocation]# - Shows an image inside the .wz files.
- #L[number]# Selection open.
- \r\n - Moves down a line.
- \r = Return Carriage
- \n = New Line
- \t = Tab (4 spaces)
- \b = Backwards

## Acknowledgements 
- [Vana](https://github.com/retep998/Vana)
- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started

## Screenshots

![](https://i.imgur.com/RIp8OWV.png)

![](https://i.imgur.com/2wYVksH.png)

![](https://i.imgur.com/g7OEhTc.png)

![](https://i.imgur.com/ovAujlt.png)

![](https://i.imgur.com/hE0mWItg.png)

![](https://i.imgur.com/4bizhIi.png)