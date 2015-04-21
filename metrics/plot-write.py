import numpy as np
import matplotlib.pyplot as plt

import csv

# calculate avg write per round (for leader and clients)
playerSizes = [1, 2, 3, 4, 5]

writeAvgsLeader = []
writeAvgsClient1 = []
#writeAvgsClient2 = []

for i in playerSizes:
	with open("writeThroughput{0}.csv".format(i)) as csvfile:
		reader = csv.reader(csvfile, delimiter=',')
		lastRound = 0
		writeSumLeader = 0
		writeSumClient1 = 0
		#writeSumClient2 = 0
		for row in reader:
			if int(row[0]) > lastRound:
				lastRound += 1
			if int(row[1]) == 1: # leader id = 1
				writeSumLeader += int(row[2])
			elif int(row[1]) == 2: # some client id
				writeSumClient1 += int(row[2])
			#elif int(row[1]) == 3: # some client id
			#	writeSumClient2 += int(row[2])


		writeAvgsLeader.append(writeSumLeader / lastRound)
		writeAvgsClient1.append(writeSumClient1 / lastRound)
		#writeAvgsClient2.append(writeSumClient2 / lastRound)

# write throughput
plt.scatter(playerSizes, writeAvgsLeader, color='r')
plt.scatter(playerSizes, writeAvgsClient1, color='b')

# determine best fit line (leader)
par = np.polyfit(playerSizes, writeAvgsLeader, 1, full=True)
slope=par[0][0]
intercept=par[0][1]
xl = [min(playerSizes), max(playerSizes)]
yl = [slope*xx + intercept for xx in xl]
plt.plot(xl, yl, '-r', label='leader')

# determine best fit line (client)
par = np.polyfit(playerSizes, writeAvgsClient1, 1, full=True)
slope=par[0][0]
intercept=par[0][1]
x2 = [min(playerSizes), max(playerSizes)]
y2 = [slope*xx + intercept for xx in x2]
plt.plot(x2, y2, '--b', label ='client')

plt.title('Write Throughput by # Players')
plt.xlabel("# Players")
plt.ylabel("Average # Bytes Written per Round")
plt.legend(loc='upper left')

plt.show()
