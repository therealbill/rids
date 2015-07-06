# RS Data Ingester

This is a data ingester for RedSkull platforms. It connects to RedSkull via the
Redskull RPC API to retrieve `INFO` data for the given instance.

# Usage
   rs-dataingester [global options] command [command options] [arguments...]

# Commands
   * ingest	Ingest known pods data
   * help, h	Shows a list of commands or help for one command
   
# Global Options
   * --rpcaddr, -r "localhost:8001"	Redskull RCP address in form 'ip:port' [$REDSKULL_RPCADDR]
   * --rsaddress, --rsa "localhost:8000"	Redskull HTTP address in form 'ip:port' [$REDSKULL_HTTPADDR]
   * --serviceaddress, -s "localhost:443"	RIDS Service address in form 'ip:port' [$RIDS_ADDRESS]
   * --help, -h				show help
   * --generate-bash-completion		
   * --version, -v			print the version
   
Each command line option used for configuration are also configurable via
environment variables as listed. This allows you to configure the ingester
entirely via the environment.
