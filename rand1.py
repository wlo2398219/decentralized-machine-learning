import random
import string
import sys
import os
import networkx as nx
import matplotlib.pyplot as plt


head = "127.0.0.1:"

p = 0.2
commands = []
template = "./finalproject -UIPort=%d -gossipAddr=%s -name=%s -peers=%s -rtimer=5 -mode=%s -byz=%s > %s &"
random.seed(sys.argv[1])

byz = ''
if len(sys.argv) > 3:
    byz = sys.argv[3]


name_addr = {}
edge_list = []

for ind, n in enumerate(string.ascii_uppercase[:10]):
    name_addr[n] = (head+str(ind+5000))

for input_name in string.ascii_uppercase[:10]:
    peers = []
    ind, UIPort, gossipPort = 0, 10000, 5000
    
    for name, port in zip(string.ascii_uppercase[:10], range(5000, 5010)):

        if random.uniform(0, 1) <= p and name != input_name:
            peers.append(head+str(port))
            edge_list.append((name, input_name))

        if name == input_name:
            UIPort += ind
            gossipPort += ind
        
        ind += 1
    
    peers = ','.join(peers)
    gossipAddr = head + str(gossipPort)
    outputFile = input_name + ".out"
    
    if input_name == byz:
        is_byz = True
    else:
        is_byz = False
    commands.append(template%(UIPort, gossipAddr, input_name, peers, sys.argv[2], is_byz, outputFile))    
    # print(commands[-1])

draw = False
if draw:
    G = nx.Graph()

    for n in string.ascii_uppercase[:10]:
        G.add_node(n)

    G.add_edges_from(edge_list)

    nx.draw_circular(G, with_labels=True,  alpha = 0.7)
    plt.savefig('topo.png', format='PNG')    

for command, n in zip(commands, string.ascii_uppercase[:10]):
    os.chdir(n)
    os.system(command)
    print(command)
    os.chdir("..")
