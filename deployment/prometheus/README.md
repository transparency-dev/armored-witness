# Prometheus Monitoring for Armored Witness

This is a simple docker setup for bringing up a local prometheus instance that will pull metrics from an armored witness running on the local network.

To deploy this:
 1. Copy this directory out of the git repo
 2. Edit `prometheus.yaml` to change the IP address to your local armored witness IP address
 3. Run `docker compose up -d`

Prometheus will now be running, collecting metrics from the armored witness.
You can access the prometheus UI at `http://localhost:9091` to investigate the metrics.
