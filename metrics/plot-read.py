import numpy as np
import matplotlib.pyplot as plt

import csv

# calculate avg read per round (for leader and clients)
playerSizes = [1, 2, 3, 4, 5, 6, 7, 8, 17]

readAvgsLeader = []
readAvgsClient1 = []
#readAvgsClient2 = []

for i in playerSizes:
	with open("readThroughput{0}.csv".format(i)) as csvfile:
		reader = csv.reader(csvfile, delimiter=',')
		lastRound = 0
		readSumLeader = 0
		readSumClient1 = 0
		#readSumClient2 = 0
		for row in reader:
			if int(row[0]) > lastRound:
				lastRound += 1
			if int(row[1]) == 1: # leader id = 1
				readSumLeader += int(row[2])
			elif int(row[1]) == 2: # some client id
				readSumClient1 += int(row[2])
			#elif int(row[1]) == 3: # some client id
			#	readSumClient2 += int(row[2])


		readAvgsLeader.append(readSumLeader / lastRound)
		readAvgsClient1.append(readSumClient1 / lastRound)
		#readAvgsClient2.append(readSumClient2 / lastRound)

plt.plot(playerSizes, readAvgsLeader, playerSizes, readAvgsClient1, marker='o')
plt.show()
