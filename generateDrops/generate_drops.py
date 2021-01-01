import json
import sqlite3
import sys

def generate(mobs_file, sql_file, drops_file):
    with open(mobs_file) as json_file:
        mobs = json.load(json_file)
    
    connection = sqlite3.connect(":memory:")
    cursor = connection.cursor()
    sql_file = open(sql_file)
    sql_as_string = sql_file.read()
    cursor.executescript(sql_as_string)

    drops = dict()

    for mob in mobs:
        drops[mob] = []
        for row in cursor.execute("SELECT * FROM dropdata WHERE dropperid='" + str(mob) + "'"):
            drops[mob].append({
                "isMesos": row[2] == 1,
                "itemId" : row[3],
                "min": row[4],
                "max": row[5],
                "questId": row[6],
                "chance": row[7],
            })
    
    with open(drops_file, 'w') as out_file:
        json.dump(drops, out_file)

if __name__=="__main__":
    if len(sys.argv) != 4:
        sys.exit(1)

    generate(sys.argv[1], sys.argv[2], sys.argv[3])
