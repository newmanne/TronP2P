import numpy as np
import matplotlib.pyplot as plt

import csv

# calculate avg read per round (for leader and clients)
playerSizes = [1, 2, 3, 4, 5, 6, 7, 8]

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

readAvgsClient1[0] = readAvgsLeader[0]

# read throughput
plt.scatter(playerSizes, readAvgsLeader, color='r')
plt.scatter(playerSizes, readAvgsClient1, color='b')

# determine best fit line (leader)
par = np.polyfit(playerSizes, readAvgsLeader, 1, full=True)
slope=par[0][0]
intercept=par[0][1]
xl = [min(playerSizes), max(playerSizes)]
yl = [slope*xx + intercept for xx in xl]
plt.plot(xl, yl, '-r', label='leader')

# determine best fit line (client)
par = np.polyfit(playerSizes, readAvgsClient1, 1, full=True)
slope=par[0][0]
intercept=par[0][1]
x2 = [min(playerSizes), max(playerSizes)]
y2 = [slope*xx + intercept for xx in x2]
plt.plot(x2, y2, '--b', label ='client')

plt.title('Read Throughput by # Players')
plt.xlabel("# Players")
plt.ylabel("Average # Bytes Read per Round")
plt.legend(loc='upper left')

plt.show()

# write throughput (TODO, for project report)

