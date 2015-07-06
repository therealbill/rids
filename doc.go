// build +ignore
package main

/*

RIDS is a service which allows you to build a Redis inventory and metrics
collection system. It can be used to keep track of Redis instances by platform
as well as other aspects.


It works by running a central API Server (or multiple running behing
round-robin DNS or a load balancer) which provides an API to input instance
connection information as well as get that data, pull metrics, and upload that
data via platform specific ingesters.

*/
