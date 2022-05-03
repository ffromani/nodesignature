# nodesignature computes the signature of a working set

A "working set" is a unordered set of namespaced work units running at any given time on a "node".
The "signature" is a unique identifer which is bound to that specific working set.
This allows to uniquely and concisely identify a workingset without enumerating all the names all
the time. This concepts maps nicely on kubernetes (but is not limited to):
"namespaced work units" = pods and "node" = (kubernetes) nodes.

## LICENSE

apache v2

