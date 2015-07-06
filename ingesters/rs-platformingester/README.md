# RedSkull PDI

This is a Platform  Ingestor for the Redis Instance Data Service. It
connects to Red Skull to learn of all pods monitored by the RS+Sentinel cluster
and upload the connection information for that  data to the RIDS API service.

# Usage
   rs-platformingester [global options] command [command options] [arguments...]


# Commands

  * ingest	Ingest known pods connection data
  * help, h	Shows a list of commands or help for one command
   
# Global Options
   * --rpcaddr, -r "localhost:8001"	Redskull RCP address in form 'ip:port' [$REDSKULL_RPCADDR]
   * --serviceaddress, -s "localhost:443"	RIDS Service address in form 'ip:port' [$RIDS_ADDRESS]
   * --help, -h				show help
   * --generate-bash-completion		
   * --version, -v			print the version
   
Each command line option used for configuration are also configurable via
environment variables as listed. This allows you to configure the ingester
entirely via the environment.
