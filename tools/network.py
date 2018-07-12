#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Jul 12 12:33:08 2018

Describes network components

@author: yaric
"""

import networkx as nx
import sys

class Node(object):
    """
    This describes network node with specific attributes
    """
    def __init__(self, node_id, neuron_type):
        """
        Creates new network node with specified id for given neuron type
        Argumets:
            node_id: the node ID
            neuron_type: the type of neuron unit presented by this node [0 - hidden, 1 - input, 2 - output, 3 - bias]
        """
        self.node_id = node_id
        self.neuron_type = neuron_type
        
    def color(self):
        """
        Returns color of this node
        """
        if self.neuron_type == 0:
            return 'white'
        elif self.neuron_type == 1:
            return 'blue'
        elif self.neuron_type == 2:
            return 'red'
        else:
            return 'yellow'


class Link(object):
    """
    Represent link between two nodes of the network with associated weight
    """
    def __init__(self, node_in, node_out, weight):
        """
        Creates new network edge connecting particular network nodes
        Argumets:
            node_in: the ID of inputing node
            node_out: the ID of node this link affects
            weight: the weight of this link
        """
        self.node_in = node_in
        self.node_out = node_out
        self.weight = weight

class Network(object):
    """
    This describes simple network consisting from nodes and links between them
    """
    def __init__(self):
        """
        Creates new network and initialize internal structures
        """
        self.nodes = {}
        self.links = []
        
    def addNode(self, node_id, neuron_type):
        """
        Adds specified node to the network
        """
        node = Node(node_id, neuron_type)
        self.nodes[node_id] = node
        
    def addLink(self, node_in, node_out, weight):
        """
        Adds specified link to the network
        """
        link = Link(node_in, node_out, weight)
        self.links.append(link)
        
    def buildGraph(self):
        """
        Builds graph from this network and returns it.
        """
        self.normalize()
        
        g = nx.DiGraph()
        
        for k, n in self.nodes.items():
            g.add_node(n.node_id, color=n.color())
            
        for l in self.links:
            if l.node_in not in self.nodes or l.node_out not in self.nodes:
                raise Exception('Incorrect link detected')
            elif l.node_in == l.node_out:
                nx.add_cycle(g, [l.node_in, l.node_in], weight=l.weight)
            else:
                g.add_edge(l.node_in, l.node_out, weight=l.weight)    
                
        return g
    
    def normalize(self, offset=0.5):
        """
        Used to perform links normalization to be in range [0, inf), i.e. any
        positive numbers.
        """
        minw = sys.float_info.max
        maxw = sys.float_info.min
        for l in self.links:
            if l.weight < minw:
                minw = l.weight
                
            if l.weight > maxw:
                maxw = l.weight
                
        v = abs(minw) + maxw
        for l in self.links:
            l.weight = offset + (l.weight + abs(minw)) / v