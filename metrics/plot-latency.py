import numpy as np
import matplotlib.pyplot as plt

import csv

def plotRoundLatency():
	rounds = []
	pids = []
	deltas = []
	with open('roundLatency.csv', 'rb') as csvfile:
		reader = csv.reader(csvfile, delimiter=',')
		for row in reader:
			rounds.append(int(row[0]))
			pids.append(int(row[1]))
			deltas.append(int(row[2]))

	# Remove bad first time
	del rounds[0]
	del pids[0]
	del deltas[0]

	# TODO: make nice with labels, fixed axis, perhaps colour according to pid (indicate leader change)
	plt.plot(rounds, deltas, marker='o')
	plt.show()


def plotReadThroughout():
	rows = []
	rounds = []
	pids = []
	tputs = []
	with open('readThroughput.csv', 'rb') as csvfile:
		reader = csv.reader(csvfile, delimiter=',')
		for row in reader:
			rows.append([int(row[0]), int(row[1]), int(row[2])])

	rows.sort()
	big = 0
	tput = 0
	for row in rows:
		if row[1] == 1:
			if row[0] > big:
				rounds.append(row[0])
				pids.append(row[1])
				tputs.append(tput) # sum up tputs by round for cumulative read tput per round?
				tput = 0
				big = row[0]
			else:
				tput += row[2]

	# TODO: plot different colours for different players?
	plt.bar(rounds, tputs)
	plt.show()





