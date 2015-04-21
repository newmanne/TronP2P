import numpy as np
import matplotlib.pyplot as plt

import csv

"""
# basic latency graph with one player TODO change scale to match others
rounds = []
pids = []
deltas = []

with open('roundLatency1.csv', 'rb') as csvfile:
	reader = csv.reader(csvfile, delimiter=',')
	for row in reader:
		rounds.append(int(row[0]))
		pids.append(int(row[1]))
		deltas.append(int(row[2]) / 10**6)

rounds = rounds[5:]
deltas = deltas[5:]
plt.plot(rounds, deltas, marker='o')
plt.title('Latency between Rounds')
plt.xlim([0, 250])
plt.ylim([0, 800])
plt.ylabel('Latency (ms)')
plt.xlabel('Round Number')
plt.show()


# latency with client drop
rounds = []
pids = []
deltas = []
with open('roundLatency2withClientDrop.csv', 'rb') as csvfile:
	reader = csv.reader(csvfile, delimiter=',')
	for row in reader:
		rounds.append(int(row[0]))
		pids.append(int(row[1]))
		deltas.append(int(row[2]) / 10**6)

rounds = rounds[5:]
deltas = deltas[5:]
plt.plot(rounds, deltas, marker='o')
plt.xlim([0, 125])
plt.ylim([0, 2500])
plt.title('Latency between Rounds (with Client drop)')
plt.ylabel('Latency (ms)')
plt.xlabel('Round Number')
plt.show()

# latency with leader drop
rounds = []
pids = []
deltas = []
with open('roundLatency2withLeaderDrop.csv', 'rb') as csvfile:
	reader = csv.reader(csvfile, delimiter=',')
	for row in reader:
		rounds.append(int(row[0]))
		pids.append(int(row[1]))
		deltas.append(int(row[2]) / 10**6)

rounds = rounds[5:500]
deltas = deltas[5:500]
plt.plot(rounds, deltas, marker='o')
plt.xlim([0, 250])
plt.ylim([0, 4500])
plt.title('Latency between Rounds (with Leader drop)')
plt.ylabel('Latency (ms)')
plt.xlabel('Round Number')
plt.show()
"""

# basic latency graph with one player TODO change scale to match others
rounds = []
pids = []
deltas = []

with open('roundLatency4LeaderDrops.csv', 'rb') as csvfile:
	reader = csv.reader(csvfile, delimiter=',')
	for row in reader:
		rounds.append(int(row[0]))
		pids.append(int(row[1]))
		deltas.append(int(row[2]) / 10**6)

rounds = rounds[5:]
deltas = deltas[5:]
plt.plot(rounds, deltas, marker='o')
plt.title('Latency between Rounds')
plt.xlim([0, 80])
plt.ylim([0, 5000])
plt.ylabel('Latency (ms)')
plt.xlabel('Round Number')
plt.show()
