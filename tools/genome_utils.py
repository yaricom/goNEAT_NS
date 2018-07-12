#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Jul 12 12:15:45 2018

Utility script to generate network graphs visualization from genome specification
files.

@author: yaric
"""
import argparse
import re

import networkx as nx

import network as n

def readGenome(path):
    """
    Reads genome from specified path and retruns it as a network of nodes and links.
    """
    net = n.Network()
    with open(path, 'r') as lines:
        for line in lines:
            if 'node' in line:
                params = [int(s) for s in re.findall(r'[-+]?[0-9]*\.?[0-9]+', line)]
                net.addNode(params[0], params[3])
            elif 'gene' in line:
                params = re.findall(r'[-+]?[0-9]*\.?[0-9]+', line)
                net.addLink(int(params[1]), int(params[2]), float(params[3]))
    
    return net
                
                    
if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='The genome file utilities')
    parser.add_argument('input_file', help='the input data file with genome encoded')
    parser.add_argument('--out', help='the file to store results')
    parser.add_argument('--operation', default = 'GraphML', help='the operation to be applied')
    
    
    args = parser.parse_args()
    if args.operation == 'GraphML':
        net = readGenome(args.input_file)
        G = net.buildGraph()
        nx.write_graphml(G, args.out)
    else:
        print("Unsupported operation requested: %s" % args.operation)
        