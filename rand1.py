import random
import string
import sys
import os
import networkx as nx
import matplotlib.pyplot as plt
import time

head = "127.0.0.1:"

p = 0.25
commands = []
template_gui = "./finalproject -UIPort=%d -gossipAddr=%s -name=%s -peers=%s -rtimer=5 -mode=%s -gui -byz=%s > %s &"
template = "./finalproject -UIPort=%d -gossipAddr=%s -name=%s -peers=%s -rtimer=5 -mode=%s -byz=%s > %s &"
random.seed(sys.argv[1])
NUM = int(sys.argv[4])

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

        # ring topology
        # if random.uniform(0, 1) <= p and name != input_name:
        #     peers.append(head+str(port))
        #     edge_list.append((name, input_name))

        # ER graph
        if random.uniform(0, 1) <= p and name != input_name:
            peers.append(head+str(port))
            edge_list.append((name, input_name))


        if name == input_name:
            UIPort += ind
            gossipPort += ind

        ind += 1

        # star topology
        # if input_name == "A":
        #     peers = []
        #     for name, port in zip(string.ascii_uppercase[1:10], range(5001, 5010)):
        #         peers.append(head+str(port))
        #         edge_list.append((name, input_name))
        # else:
        #     peers = []
        
    peers = ','.join(peers)
    gossipAddr = head + str(gossipPort)
    outputFile = input_name + ".out"
    
    if input_name == byz:
        is_byz = True
    else:
        is_byz = False

    if input_name == "A":
        commands.append(template_gui%(UIPort, gossipAddr, input_name, peers, sys.argv[2], is_byz, outputFile))    
    else:
        commands.append(template%(UIPort, gossipAddr, input_name, peers, sys.argv[2], is_byz, outputFile))    
    # print(commands[-1])

draw = True
if draw:
    G = nx.Graph()

    for n in string.ascii_uppercase[:10]:
        G.add_node(n)

    G.add_edges_from(edge_list)

    nx.draw_circular(G, with_labels=True,  alpha = 0.7)
    plt.savefig('topo.png', format='PNG')    

for command, n in zip(commands, string.ascii_uppercase[:NUM]):
    os.chdir(n)
    os.system(command)
    print(command)
    os.chdir("..")
