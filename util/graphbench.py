import numpy as np
import matplotlib.pyplot as plt
import glob,csv
import random

benchmarkfiles = glob.glob("*.bench")

benchmarks = {}

#graphtype = "nanoseconds"
graphtype = "ops_per_second"


for fname in benchmarkfiles:
    testname = fname[:-6]
    print "TEST",testname
    f = open(fname,"r")
    

    reader = csv.reader(f,delimiter="\t")
    for row in reader:
        benchname = row[0][9:]
        if not benchname in benchmarks:
            benchmarks[benchname] = {}
        nanoseconds = float(row[2].strip().split(' ')[0])
        val = None
        if graphtype == "nanoseconds":
            val = nanoseconds
        elif graphtype == "ops_per_second":
            val = 1000000000./nanoseconds
        benchmarks[benchname][testname] = val

    f.close()



#Generate arrays for each benchmark
nameArray = []
benchValues = []
benchNames = []

#First get the nameArray and benchValues initialized
k = benchmarks.keys()[0]
for name in benchmarks[k]:
    nameArray.append(name)
    benchValues.append([])

for k in benchmarks:
    benchNames.append(k)
    for i in xrange(len(nameArray)):
        benchValues[i].append(benchmarks[k][nameArray[i]])

print nameArray
print benchValues
print benchNames
ind = np.arange(len(benchNames))  # the x locations for the groups
width = 0.7/len(nameArray)       # the width of the bars

fig, ax = plt.subplots()

bars = []
b1 = []

for i in range(len(nameArray)):
    b = ax.bar(ind+i*width,benchValues[i],width,color=np.random.rand(3,1))
    bars.append(b)
    b1.append(b)

ax.set_title("Benchmarks of ConnectorDB")
ax.set_xticks(ind+0.35)
ax.set_xticklabels(benchNames,rotation='vertical')
ax.legend(b1,nameArray)


#Label the values
for bar in bars:
    for rect in bar:
        height = rect.get_height()
        ax.text(rect.get_x()+rect.get_width()/2.,1.05*height,'%d'%int(height),ha='center',va='bottom')

ax.set_yscale('log')

if graphtype == "nanoseconds":
    ax.set_ylabel("ns/op")
elif graphtype == "ops_per_second":
    ax.set_ylabel("ops/s")



plt.show()
