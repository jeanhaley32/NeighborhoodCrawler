# INCOMPLETE

## NeighborHood Crawler

Crawls the ethereum network, finding nodes and appending a list of neighbors to each. 

This is a very early Proof of Concept. And it currently is not functioning.

### How it works

The ```devp2p discv4 crawl``` command will produce a JSON file of all of the nodes it finds. 
So far this script does

  1. Injest that JSON list of Nodes
  2. Crawl through each list, gathering it's neighbors
  3. Appends that list of neighbors to a new "neighbors" field
  4. Print out the modified json list.

I was able to create a version of this code that worked when fed one node. But for some reason i'm unable to get he same results here.
This still needs debugging.

### Notes

This is a very simple proof of concept.
A future iteration of this concept would be used to:

  1. Use parallelism (goroutines) to traverse the list faster.
  2. Combine the Crawler functionality of the devp2p library to connect the pipeline of Node traversal to neighborhood association.
     This should be two seperate scripts, one running on the output of the other.

  A Final Iterations may be more complex: \
    A potential rebuild of the whole crawler concept, if I can discover the node, and associate it's neighbors in a struct at the same time
    then having two seperate operations is unnecessary. This would look like a perpetually running service, updating node information continually. 
    ensuring neighboorhood associations are up to date, and clearing any nodes that fall off of the node map, along with removing that node assocation
    from other nodoes.

  This would most likely look like a series of goroutines, all working to perform different functions. One that regularly checks for errant nodes that
  should be removed. Maybe a "graveyard" concept, where a routine starts up, scanes all nodes in the DB for neighbors matching what is in the graveyard
  popping those nodes off of the list.

  There are a number of different concepts that could be introduced. Including layers that work on purely gathering statistics and exporting them as digestible
  metrics for third party tool to then ingest.
