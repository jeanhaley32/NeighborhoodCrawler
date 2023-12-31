# NeighborHood Crawler

## 🚧 **Work in Progress** 🚧

>Crawls the ethereum network, finding nodes, and appending a list of neighbors to each

This is a very early Proof of Concept. And it currently is not functioning.

### How it works

The ```devp2p discv4 crawl``` command will produce a JSON file of all of the nodes it finds.
So far this script...

  1. Ingests that JSON list of Nodes
  2. Crawls through each list, gathering it's neighbors
  3. Appends that list of neighbors to a new "neighbors" field
  4. Prints out the modified json list.

I was able to create a version of this code that worked when fed one node. But for some reason i'm unable to get those same results here.
This still needs debugging.

### Notes

This is a very simple proof of concept.
A future iteration of this POC would:

  1. Use parallelism (goroutines) to traverse the list faster.
  2. Fold in the Crawler functionality of the devp2p library to combine the pipeline of [DHT](https://en.wikipedia.org/wiki/Distributed_hash_table) Traversal and Node Discovery with neighborhood association.
     This should not be two separate scripts, one running on the output of the other, but instead one unified process.

  A Final Iterations may be more complex: \
    A potential rebuild of the whole crawler concept, if I can discover the node, and associate it's neighbors in a struct at the same time
    then having two separate operations is unnecessary. This would look like a perpetually running service, updating node information continually.
    ensuring neighborhood associations are up to date, and clearing any nodes that fall off of the node map, along with removing that node association
    from other nodes.

  This would most likely look like a series of goroutines, all working to perform different functions. One that regularly checks for errant nodes that
  should be removed. Maybe a "graveyard" concept, where a routine starts up, scans all nodes in the DB for neighbors matching what is in the graveyard
  popping those nodes off of the list.

  There are a number of different concepts that could be introduced. Including layers that work on purely gathering statistics and exporting them as digestible
  metrics for third party tool to then ingest.
